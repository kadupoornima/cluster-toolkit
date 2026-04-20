// Copyright 2026 "Google LLC"
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telemetry

import (
	"context"
	"fmt"
	"slices"
	"strings"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	resourcemanagerpb "cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"

	"hpc-toolkit/pkg/shell"

	"hpc-toolkit/pkg/config"
	"hpc-toolkit/pkg/logging"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zclconf/go-cty/cty"
)

var (
	isGkeModulePatterns        = []string{".*gke-node-pool.*", ".*gke-cluster.*"}
	isSlurmModulePatterns      = []string{".*schedmd-slurm-gcp-.*"}
	IsVmInstanceModulePatterns = []string{".*vm-instance.*"}

	machineTypeModulePattern = ".*modules.compute.*"
)

// NewCollector creates and initializes a new Telemetry Collector.
func NewCollector(cmd *cobra.Command, args []string) *Collector {
	return &Collector{
		eventCmd:       cmd,
		eventArgs:      args,
		eventStartTime: time.Now(),
		blueprint:      getBlueprint(args),
		metadata:       make(map[string]string),
	}
}

// Main function for collecting Telemetry metrics.
func (c *Collector) CollectMetrics(errorCode int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	bpModulesList := getModulesList(c.blueprint)

	c.metadata[COMMAND_FLAGS] = getCmdFlags(c.eventCmd)
	c.metadata[BLUEPRINT] = getBlueprintName(c.blueprint)
	c.metadata[DEPLOYMENT_FILE] = getDeploymentFile(c.eventCmd)
	c.metadata[IS_GKE] = getIsGke(bpModulesList)
	c.metadata[IS_SLURM] = getIsSlurm(bpModulesList)
	c.metadata[IS_VM_INSTANCE] = getIsVmInstance(bpModulesList)
	c.metadata[MACHINE_TYPE] = getMachineType(c.blueprint)
	c.metadata[REGION] = getRegion(c.blueprint)
	c.metadata[ZONE] = getZone(c.blueprint)
	c.metadata[PROVISIONING_MODE] = getProvisioningMode()
	c.metadata[MODULES] = getModules(bpModulesList)
	c.metadata[OS_NAME] = getOSName()
	c.metadata[OS_VERSION] = getOSVersion()
	c.metadata[TERRAFORM_VERSION] = getTerraformVersion()
	c.metadata[BILLING_ACCOUNT_ID] = getBillingAccountId(c.blueprint)
	c.metadata[DEPLOYED_FROM_SOURCE] = getDeployedFromSource()
	c.metadata[DEPLOYED_FROM_BINARY] = getDeployedFromBinary(c.metadata[DEPLOYED_FROM_SOURCE] == "true")
	c.metadata[IS_TEST_DATA] = getIsTestData()
	c.metadata[EXIT_CODE] = strconv.Itoa(errorCode)
}

// Method to collect Concord metrics and build event.
func (c *Collector) BuildConcordEvent() ConcordEvent {
	c.mu.Lock()
	defer c.mu.Unlock()

	return ConcordEvent{
		ConsoleType:      CLUSTER_TOOLKIT,
		EventType:        "gclusterCLI",
		EventName:        getCommandName(c.eventCmd),
		EventMetadata:    getEventMetadataKVPairs(c.metadata),
		ProjectNumber:    getProjectNumber(c.blueprint),
		ClientInstallId:  getClientInstallId(),
		BillingAccountId: getBillingAccountId(c.blueprint),
		IsGoogler:        getIsGoogler(),
		ReleaseVersion:   getReleaseVersion(),
		LatencyMs:        getLatencyMs(c.eventStartTime),
	}
}

/** Private functions **/

func getClientInstallId() string {
	return config.GetPersistentUserId()
}

func getProjectNumber(bp config.Blueprint) string {
	ctx := context.Background()
	projectID := getProjectId(bp)

	client, _ := resourcemanager.NewProjectsClient(ctx)
	defer client.Close()

	req := &resourcemanagerpb.GetProjectRequest{
		Name: fmt.Sprintf("projects/%s", projectID),
	}

	project, err := client.GetProject(ctx, req)

	if err != nil || project == nil || project.Name == "" {
		return ""
	} else {
		return strings.TrimPrefix(project.Name, "projects/")
	}
}

func getReleaseVersion() string {
	return config.GetToolkitVersion()
}

func getCommandName(cmd *cobra.Command) string {
	path := cmd.CommandPath() // Returns the full command path (e.g., "gcluster job cancel")

	if path == "" {
		return path
	} else {
		return strings.TrimPrefix(path, "gcluster ")
	}
}

func getCmdFlags(cmd *cobra.Command) string {
	flags := make([]string, 0)
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flags = append(flags, f.Name)
	})
	return strings.Join(flags, ",")
}

func getBlueprintName(bp config.Blueprint) string {
	return bp.BlueprintName
}

func getDeploymentFile(cmd *cobra.Command) string {
	var path string
	if flag := cmd.Flag("deployment-file"); flag != nil && flag.Value.String() != "" {
		path = flag.Value.String()
	}
	// else if len(args) > 0 {
	// 	// 2. For standard commands (create, deploy, destroy), the blueprint or
	// 	// deployment directory is usually the first positional argument
	// 	path = args[0]
	// }
	logging.Info("\n\n\n path: %v\n\n\n", path)
	if path != "" {
		// return filepath.Base(path)
		return path
	}
	return ""
}

func getIsGke(modulesList []string) string {
	return ifModulesMatchPatterns(modulesList, isGkeModulePatterns)
}

func getIsSlurm(modulesList []string) string {
	return ifModulesMatchPatterns(modulesList, isSlurmModulePatterns)
}

func getIsVmInstance(modulesList []string) string {
	return ifModulesMatchPatterns(modulesList, IsVmInstanceModulePatterns)
}

func getMachineType(bp config.Blueprint) string {
	var machineTypes []string
	seen := make(map[string]bool) // To keep track of added machine types to avoid duplication
	modules := getModulesWithPattern(machineTypeModulePattern, bp)

	evalAndAdd := func(key string, m config.Module) {
		if m.Settings.Has(key) {
			keyValue := m.Settings.Get(key)
			// Evaluate the value to resolve expressions like $(vars.key)
			evaluatedKey, err := bp.Eval(keyValue)
			if err != nil {
				return
			}
			// Some module outputs or references carry cty marks, so we unmark them safely before use.
			unmarkedKey, _ := evaluatedKey.Unmark()
			if !unmarkedKey.IsNull() && unmarkedKey.Type() == cty.String {
				mType := unmarkedKey.AsString()
				if !seen[mType] {
					machineTypes = append(machineTypes, mType)
					seen[mType] = true
				}
			}
		}
	}

	for _, m := range modules {
		evalAndAdd("machine_type", m)
		evalAndAdd("node_type", m) // For schedmd-slurm-gcp-v6-nodeset-tpu module. It uses node_type setting instead of machine_type.
	}
	return strings.Join(machineTypes, ",")
}

func getRegion(bp config.Blueprint) string {
	val, err := bp.Eval(config.GlobalRef("region").AsValue())
	if err == nil {
		region, _ := val.Unmark()
		if !region.IsNull() && region.Type() == cty.String {
			return region.AsString()
		}
	}
	return ""
}

func getZone(bp config.Blueprint) string {
	val, err := bp.Eval(config.GlobalRef("zone").AsValue())
	if err == nil {
		zone, _ := val.Unmark()
		if !zone.IsNull() && zone.Type() == cty.String {
			return zone.AsString()
		}
	}
	return ""
}

func getProvisioningMode() string {
	return "test"
}

// // getProvisioningMode extracts the unique consumption options (spot, reservation, on-demand)
// // used across all compute modules in the provided blueprint.
// func getProvisioningMode(bp config.Blueprint) string {
// 	// Use a map as a set to keep track of unique modes
// 	modesSet := make(map[string]bool)
// 	hasCompute := false
// 	allComputeSpotOrRes := true

// 	const (
// 		ModeSpot        = "spot"
// 		ModeReservation = "reservation"
// 		ModeOnDemand    = "on-demand"
// 	)

// 	// Iterate through all modules in the blueprint
// 	for _, mod := range config.GetAllModules(&bp) {
// 		source := strings.ToLower(mod.Source)

// 		// We are primarily interested in compute-related modules
// 		isCompute := strings.Contains(source, "compute_engine") ||
// 			strings.Contains(source, "slurm") ||
// 			strings.Contains(source, "gke_node_pool") ||
// 			strings.Contains(source, "batch")

// 		if isCompute {
// 			hasCompute = true
// 			isSpot := false
// 			isRes := false

// 			settings := mod.Settings

// 			// 1. Check for Spot / Preemptible indicators
// 			if settings.Has("spot") {
// 				val, _ := settings.Get("spot").UnmarkDeep()
// 				if val.Type() == cty.Bool && val.True() {
// 					isSpot = true
// 				}
// 			}
// 			if settings.Has("preemptible") {
// 				val, _ := settings.Get("preemptible").UnmarkDeep()
// 				if val.Type() == cty.Bool && val.True() {
// 					isSpot = true
// 				}
// 			}
// 			if settings.Has("enable_spot_vm") {
// 				val, _ := settings.Get("enable_spot_vm").UnmarkDeep()
// 				if val.Type() == cty.Bool && val.True() {
// 					isSpot = true
// 				}
// 			}
// 			if settings.Has("provisioning_model") {
// 				val, _ := settings.Get("provisioning_model").UnmarkDeep()
// 				if val.Type() == cty.String && strings.ToUpper(val.AsString()) == "SPOT" {
// 					isSpot = true
// 				}
// 			}

// 			// 2. Check for Reservation indicators
// 			if settings.Has("reservation_affinity") {
// 				isRes = true
// 			}
// 			if settings.Has("reservation_name") {
// 				val, _ := settings.Get("reservation_name").UnmarkDeep()
// 				if val.Type() == cty.String && val.AsString() != "" {
// 					isRes = true
// 				}
// 			}
// 			if settings.Has("consume_reservation_type") {
// 				val, _ := settings.Get("consume_reservation_type").UnmarkDeep()
// 				if val.Type() == cty.String && val.AsString() != "" && val.AsString() != "NO_RESERVATION" {
// 					isRes = true
// 				}
// 			}

// 			// Record the modes found for this module
// 			if isSpot {
// 				modesSet[ModeSpot] = true
// 			}
// 			if isRes {
// 				modesSet[ModeReservation] = true
// 			}

// 			// If it's a compute module but uses neither spot nor reservations, it falls back to on-demand
// 			if !isSpot && !isRes {
// 				allComputeSpotOrRes = false
// 			}
// 		}
// 	}

// 	// Add "on-demand" if there was at least one compute module that wasn't strictly spot or reservation
// 	if hasCompute && !allComputeSpotOrRes {
// 		modesSet[ModeOnDemand] = true
// 	}

// 	// Convert the set to a slice
// 	var modes []string
// 	for mode := range modesSet {
// 		modes = append(modes, mode)
// 	}

// 	return strings.Join(modes, ",")
// }

// func getModules(modulesList []string) string {
// 	return strings.Join(modulesList, ",")
// }

func getModules(modulesList []string) string {
	sanitizedModules := make([]string, 0)
	standardModules, _ := config.GetAllDefinedModules()
	for _, m := range modulesList {
		if slices.Contains(standardModules, m) {
			sanitizedModules = append(sanitizedModules, m)
		} else {
			sanitizedModules = append(sanitizedModules, "Custom module")
		}
	}
	return strings.Join(sanitizedModules, ",")
}

// func logModule(module config.Module) {
// 	logging.Info("MMM: Source: %v", module.Source)
// 	logging.Info("MMM: Kind: %v", module.Kind)
// 	logging.Info("MMM: ID: %v", module.ID)
// 	logging.Info("MMM: Use: %v", module.Use)
// 	logging.Info("MMM: Outputs: %v", module.Outputs)
// 	logging.Info("MMM: Settings: %v", module.Settings)
// 	logging.Info("\n\n\n")
// }

func getOSName() string {
	return runtime.GOOS
}

// getOSVersion returns the OS version of the current system.
func getOSVersion() string {
	switch runtime.GOOS {
	case "linux":
		return getLinuxVersion()
	case "darwin":
		return getMacVersion()
	case "windows":
		return getWindowsVersion()
	default:
		return ""
	}
}

func getTerraformVersion() string {
	version, err := shell.TfVersion()
	if err != nil {
		return "Unknown"
	}
	return version
}

// getDeployedFromSource returns "true" if the CLI is being run from inside a Git repository,
// indicating the user likely cloned the code and built it locally.
func getDeployedFromSource() string {
	exePath, err := os.Executable()
	if err != nil {
		return "false"
	}
	exeDir := filepath.Dir(exePath)
	gitPath := filepath.Join(exeDir, ".git")

	// Check if the .git folder exists in the same directory.
	if _, err := os.Stat(gitPath); err == nil {
		return "true"
	}

	return "false"
}

// If not deployed from source, deployed from binary.
func getDeployedFromBinary(deployedFromSource bool) string {
	return fmt.Sprintf("%v", !deployedFromSource)
}

// This method intentionally returns "true", as all current telemetry is in testing phase.
func getIsTestData() string {
	return "true" // do not modify
}

func getBillingAccountId(bp config.Blueprint) string {
	projectID := getProjectId(bp)
	if projectID == "" {
		return ""
	}

	ctx := context.Background()
	billingAccount := getProjectBillingAccount(ctx, projectID)
	if billingAccount == "" {
		return ""
	} else {
		return strings.TrimPrefix(billingAccount, "billingAccounts/")
	}
}

// getIsGoogler identifies if the CLI is being run by an internal Google user.
func getIsGoogler() bool {
	return isInternalUser() || hasProdAccess()
}

func getLatencyMs(eventStartTime time.Time) int64 {
	return time.Since(eventStartTime).Milliseconds()
}
