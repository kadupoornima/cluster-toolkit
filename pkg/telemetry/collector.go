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
	"crypto/sha256"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	resourcemanagerpb "cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"

	"hpc-toolkit/pkg/shell"

	"hpc-toolkit/pkg/config"
	"hpc-toolkit/pkg/logging"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	crm "google.golang.org/api/cloudresourcemanager/v1"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	isGkeModulePatterns        = []string{".*gke-node-pool.*", ".*gke-cluster.*"}
	isSlurmModulePatterns      = []string{".*schedmd-slurm-gcp-.*"}
	IsVmInstanceModulePatterns = []string{".*vm-instance.*"}
	machineTypeModulePattern   = ".*modules.compute.*"
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

func (c *Collector) CollectMetrics(errorCode int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.modulesList = getModulesList(c.blueprint)

	c.metadata[COMMAND_FLAGS] = getCmdFlags(c.eventCmd)
	c.metadata[BLUEPRINT] = getBlueprintName(c.blueprint)
	c.metadata[DEPLOYMENT_FILE] = getDeploymentFile(c.eventCmd)
	c.metadata[IS_GKE] = getIsGke(c.modulesList)
	c.metadata[IS_SLURM] = getIsSlurm(c.modulesList)
	c.metadata[IS_VM_INSTANCE] = getIsVmInstance(c.modulesList)
	c.metadata[MACHINE_TYPE] = getMachineType(c.blueprint)
	c.metadata[REGION] = getRegion(c.blueprint)
	c.metadata[ZONE] = getZone(c.blueprint)
	c.metadata[PROVISIONING_MODE] = getProvisioningMode()
	c.metadata[MODULES] = getModules(c.modulesList)
	c.metadata[OS_NAME] = getOSName()
	c.metadata[OS_VERSION] = getOSVersion()
	c.metadata[TERRAFORM_VERSION] = getTerraformVersion()
	c.metadata[DEPLOYED_FROM_SOURCE] = getDeployedFromSource()
	c.metadata[DEPLOYED_FROM_BINARY] = getDeployedFromBinary()
	c.metadata[IS_TEST_DATA] = getIsTestData()
	c.metadata[EXIT_CODE] = strconv.Itoa(errorCode)
	c.metadata[BILLING_ACCOUNT] = getBillingAccount(c.blueprint)
}

func (c *Collector) BuildConcordEvent() ConcordEvent {
	c.mu.Lock()
	defer c.mu.Unlock()

	return ConcordEvent{
		ConsoleType:     CLUSTER_TOOLKIT,
		EventType:       "gclusterCLI",
		EventName:       getCommandName(c.eventCmd),
		EventMetadata:   getEventMetadataKVPairs(c.metadata),
		LatencyMs:       getLatencyMs(c.eventStartTime),
		ProjectNumber:   getProjectNumber(c.blueprint),
		ClientInstallId: getClientInstallId(),
		IsGoogler:       getIsGoogler(c.blueprint),
		ReleaseVersion:  getReleaseVersion(),
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

	// The Name field in the request can be the Project ID
	req := &resourcemanagerpb.GetProjectRequest{
		Name: fmt.Sprintf("projects/%s", projectID),
	}

	project, err := client.GetProject(ctx, req)
	if err != nil {
		logging.Error("Could not get project: %v", err)
	}

	// project.Name returns "projects/123456789012"
	projectNumber := strings.TrimPrefix(project.Name, "projects/")

	fmt.Printf("Project ID: %s\n", project.ProjectId)
	fmt.Printf("Project Number: %s\n", projectNumber)
	return projectNumber
}

func getReleaseVersion() string {
	return config.GetToolkitVersion()
}

func getCommandName(cmd *cobra.Command) string {
	return cmd.Name()
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
	isGke := false
	isGke = slices.ContainsFunc(modulesList, func(s string) bool {
		for _, pattern := range isGkeModulePatterns {
			match, _ := regexp.MatchString(pattern, s)
			isGke = isGke || match
		}
		return isGke
	})
	return strconv.FormatBool(isGke)
}

func getIsSlurm(modulesList []string) string {
	isSlurm := false
	isSlurm = slices.ContainsFunc(modulesList, func(s string) bool {
		for _, pattern := range isSlurmModulePatterns {
			match, _ := regexp.MatchString(pattern, s)
			isSlurm = isSlurm || match
		}
		return isSlurm
	})
	return strconv.FormatBool(isSlurm)
}

func getIsVmInstance(modulesList []string) string {
	isSlurm := false
	isSlurm = slices.ContainsFunc(modulesList, func(s string) bool {
		for _, pattern := range IsVmInstanceModulePatterns {
			match, _ := regexp.MatchString(pattern, s)
			isSlurm = isSlurm || match
		}
		return isSlurm
	})
	return strconv.FormatBool(isSlurm)
}

func getMachineType(bp config.Blueprint) string {
	machine_types := make([]any, 0)
	modules := getModulesWithPattern(machineTypeModulePattern, bp)
	// logging.Info("\nMODULES:\n%v\n", modules)

	for _, m := range modules {
		if m.Settings.Has("machine_type") {
			logging.Info("\nBBBBB: %v", m.Settings.Get("machine_type"))
			machine_type := m.Settings.Get("machine_type")
			logging.Info("machine_type.Type(): %v", machine_type.Type())
			// logging.Info("BBBBB: %v\n", machine_type.AsValueMap())

			if !machine_type.IsNull() {
				machine_types = append(machine_types, machine_type)
			}
		}
		// For the schedmd-slurm-gcp-v6-nodeset-tpu module
		// if m.Settings.Has("node_type") {
		// 	machine_type, _ := m.Settings.Get("node_type").UnmarkDeep()
		// 	if !machine_type.IsNull() && machine_type.Type() == cty.String {
		// 		machine_types = append(machine_types, machine_type.AsString())
		// 	}
		// }
	}
	logging.Info("FINAL:\n\n%v\n\n", machine_types...)
	// return strings.Join(machine_types, ",")
	return "test"
}

func getRegion(bp config.Blueprint) string {
	if bp.Vars.Has("region") {
		return bp.Vars.Get("region").AsString()
	}
	return ""
}

func getZone(bp config.Blueprint) string {
	if bp.Vars.Has("zone") {
		return bp.Vars.Get("zone").AsString()
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

func getModules(modulesList []string) string {
	return strings.Join(modulesList, ",")
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
		logging.Error("Unable to get Terraform version, %v", err)
	}
	return version
}

func getDeployedFromSource() string {
	return "test"
}

func getDeployedFromBinary() string {
	return "test"
}

func getIsTestData() string {
	return "true"
}

func getBillingAccount(bp config.Blueprint) string {
	projectID := getProjectId(bp)
	ctx := context.Background()
	billingAccount, err := getProjectBillingAccount(ctx, projectID)
	if err != nil {
		fmt.Printf("Warning: Could not fetch billing account: %v\n", err)
	} else if billingAccount == "" {
		fmt.Printf("Project %s does not have an associated billing account.\n", projectID)
	}
	billingAccount = strings.TrimPrefix(billingAccount, "billingAccounts/")
	// Hash the billing account ID to avoid PII.
	billingAccountHash := sha256.Sum256([]byte(billingAccount))
	return fmt.Sprintf("%x", billingAccountHash)[:24]
	// return billingAccount
}

// getIsGoogler returns "true" if the GCP project belongs to the Google.com organization.
func getIsGoogler(bp config.Blueprint) bool {
	// googleOrgID is the canonical Google.com organization ID
	googleOrgID := "433637338589"

	projectID := getProjectId(bp)
	if projectID == "" {
		return false
	}
	ctx := context.Background()
	service, err := crm.NewService(ctx)
	if err != nil {
		return false
	}

	// Fetch the ancestry of the project
	req := &crm.GetAncestryRequest{}
	resp, err := service.Projects.GetAncestry(projectID, req).Do()
	if err != nil {
		// This can fail if the user lacks IAM permissions or the project doesn't exist
		return false
	}

	// Traverse the ancestors from bottom (the project) to top (the organization)
	for _, ancestor := range resp.Ancestor {
		if ancestor.ResourceId.Type == "organization" && ancestor.ResourceId.Id == googleOrgID {
			return true
		}
	}

	return false
}

func getLatencyMs(eventStartTime time.Time) int64 {
	eventEndTime := time.Now()
	return eventEndTime.Sub(eventStartTime).Milliseconds()
}
