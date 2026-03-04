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
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"

	"hpc-toolkit/pkg/config"
	"hpc-toolkit/pkg/logging"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	billing "cloud.google.com/go/billing/apiv1"
	"cloud.google.com/go/billing/apiv1/billingpb"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	metadata       = make(map[string]string)
	bp             config.Blueprint
	eventStartTime time.Time
)

func CollectPreMetrics(cmd *cobra.Command, args []string) {
	eventStartTime = time.Now()
	bp = getBlueprint(args)
	logBlueprint(bp)

	metadata[USER_ID] = getUserId()
	metadata[COMMAND_NAME] = getCommandName(cmd)
	metadata[COMMAND_FLAGS] = getCmdFlags(cmd)
	metadata[BLUEPRINT] = getBlueprintName(bp)
	metadata[DEPLOYMENT_FILE] = getDeploymentFile()
	metadata[BILLING_ACCOUNT] = getBillingAccount(bp)
	metadata[IS_GKE] = getIsGke(bp)
	metadata[IS_SLURM] = getIsSlurm(bp)
	metadata[IS_VM_INSTANCE] = getIsVmInstance()
	metadata[MACHINE_TYPE] = getMachineType(bp)
	metadata[REGION] = getRegion(bp)
	metadata[ZONE] = getZone(bp)
	metadata[PROVISIONING_MODE] = getProvisioningMode()
	metadata[MODULES] = getModules(bp)
	metadata[OS_NAME] = getOSName()
	metadata[OS_VERSION] = getOSVersion()
	metadata[TERRAFORM_VERSION] = getTerraformVersion(bp)
	metadata[IS_INTERNAL_USER] = getIsInternalUser()
	metadata[DEPLOYED_FROM_SOURCE] = getDeployedFromSource()
	metadata[DEPLOYED_FROM_BINARY] = getDeployedFromBinary()
	metadata[IS_TEST_DATA] = getIsTestData()

}

func CollectPostMetrics(errorCode int) {
	metadata[RUNTIME_MS] = getRuntime()
	metadata[EXIT_CODE] = strconv.Itoa(errorCode)
}

func getUserId() string {
	return config.GetPersistentUserId()
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

func getBlueprint(args []string) config.Blueprint {
	bp, _, _ := config.NewBlueprint(args[0])
	return bp
}
func getDeploymentFile() string {

	return "test"
}

func getModulesFromPattern(pattern string, bp config.Blueprint) []config.Module {
	modules := make([]config.Module, 0)
	for _, m := range config.GetAllModules(&bp) {
		matched, _ := regexp.Match(pattern, []byte(m.Source))
		if matched {
			modules = append(modules, m)
		}
	}
	return modules
}

func getModulesList(bp config.Blueprint) []string {
	moduleInfos := make([]config.Module, 0)
	modules := make([]string, 0)
	moduleInfos = append(moduleInfos, config.GetAllModules(&bp)...)
	for _, module := range moduleInfos {
		modules = append(modules, string(module.Source))
	}
	return modules
}

var (
	IsGkeModulePatterns           = []string{".*gke-node-pool.*", ".*gke-cluster.*"}
	IsSlurmModulePatterns         = []string{".*schedmd-slurm-gcp-.*"}
	GkeMachineTypeModulePattern   = ".*gke-node-pool.*"
	SlurmMachineTypeModulePattern = ".*schedmd-slurm-gcp-.*-nodeset.*"
)

func getIsGke(bp config.Blueprint) string {
	modules := getModulesList(bp)
	isGke := false
	isGke = slices.ContainsFunc(modules, func(s string) bool {
		for _, pattern := range IsGkeModulePatterns {
			match, _ := regexp.MatchString(pattern, s)
			isGke = isGke || match
		}
		return isGke
	})
	return strconv.FormatBool(isGke)
}

func getIsSlurm(bp config.Blueprint) string {
	modules := getModulesList(bp)
	isSlurm := false
	isSlurm = slices.ContainsFunc(modules, func(s string) bool {
		for _, pattern := range IsSlurmModulePatterns {
			match, _ := regexp.MatchString(pattern, s)
			isSlurm = isSlurm || match
		}
		return isSlurm
	})
	return strconv.FormatBool(isSlurm)
}

func getIsVmInstance() string {

	return "test"
}

func getMachineType(bp config.Blueprint) string {
	machine_types := make([]string, 0)
	modules := getModulesFromPattern(GkeMachineTypeModulePattern, bp)
	modules = append(modules, getModulesFromPattern(SlurmMachineTypeModulePattern, bp)...)
	for _, m := range modules {
		machine_types = append(machine_types, m.Settings.Get("machine_type").AsString())
	}
	return strings.Join(machine_types, ",")
}

func getProjectId(bp config.Blueprint) string {
	if bp.Vars.Has("project_id") {
		logging.Info("YYYY: %v", bp.Vars.Get("project_id"))
		return bp.Vars.Get("project_id").AsString()
		// return ""
	}
	return ""
}

// GetProjectBillingAccount fetches the billing account associated with a given GCP project in the format "billingAccounts/{billing_account_id}". If billing is disabled for the project, this will return an empty string.
func GetProjectBillingAccount(ctx context.Context, projectID string) (string, error) {
	client, err := billing.NewCloudBillingClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create billing client: %w", err)
	}
	defer client.Close()

	req := &billingpb.GetProjectBillingInfoRequest{
		Name: fmt.Sprintf("projects/%s", projectID),
	}

	info, err := client.GetProjectBillingInfo(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get billing info for project %s: %w", projectID, err)
	}
	return info.GetBillingAccountName(), nil
}

func getBillingAccount(bp config.Blueprint) string {
	projectID := getProjectId(bp)
	ctx := context.Background()
	billingAccount, err := GetProjectBillingAccount(ctx, projectID)
	if err != nil {
		fmt.Printf("Warning: Could not fetch billing account: %v\n", err)
	} else if billingAccount == "" {
		fmt.Printf("Project %s does not have an associated billing account.\n", projectID)
	}
	billingAccount = strings.TrimPrefix(billingAccount, "billingAccounts/")
	// Hash the billing account ID to avoid PII.
	billingAccountHash := sha256.Sum256([]byte(billingAccount))
	return fmt.Sprintf("%x", billingAccountHash)[:24]
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
func getIsInternalUser() string {

	return "test"
}
func getTerraformVersion(bp config.Blueprint) string {
	// tfProviders := bp.TerraformProviders

	return strconv.Itoa(len(bp.TerraformProviders))
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

func getModules(bp config.Blueprint) string {
	// logging.Info("\n\n\n")
	// moduleInfos := make([]config.Module, 0)
	// moduleInfos = append(moduleInfos, config.GetAllModules(&bp)...)
	// for _, module := range moduleInfos {
	// 	logging.Info("XXX: Source: %v", module.Source)
	// 	logging.Info("XXX: Kind: %v", module.Kind)
	// 	logging.Info("XXX: ID: %v", module.ID)
	// 	logging.Info("XXX: Use: %v", module.Use)
	// 	logging.Info("XXX: Outputs: %v", module.Outputs)
	// 	logging.Info("XXX: Settings: %v", module.Settings)
	// }
	// logging.Info("\n\n\n")

	return strings.Join(getModulesList(bp), ",")
}

func getBlueprintName(bp config.Blueprint) string {
	return bp.BlueprintName
}

func getRuntime() string {
	eventEndTime := time.Now()
	return strconv.FormatInt(eventEndTime.Sub(eventStartTime).Milliseconds(), 10)
}

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

// getLinuxVersion parses /etc/os-release to find the pretty name or version ID.
func getLinuxVersion() string {
	// Standard way to identify Linux distribution version
	f, err := os.Open("/etc/os-release")
	if err != nil {
		logging.Error("failed to open /etc/os-release: %v", err)
		return ""
	}
	defer f.Close()

	var prettyName, versionID string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName = parseOsReleaseField(line)
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = parseOsReleaseField(line)
		}
	}

	if prettyName != "" {
		return prettyName
	}
	if versionID != "" {
		return versionID
	}
	return "Linux (unknown version)"
}

// getMacVersion uses sw_vers to get the macOS product version.
func getMacVersion() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		logging.Error("sw_vers failed: %v", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getWindowsVersion uses the ver command to get the Windows version.
func getWindowsVersion() string {
	cmd := exec.Command("cmd", "/c", "ver")
	out, err := cmd.Output()
	if err != nil {
		logging.Error("ver failed: %v", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

// parseOsReleaseField helper to clean up quotes from /etc/os-release values
func parseOsReleaseField(line string) string {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.Trim(parts[1], `"`)
}

func logBlueprint(bp config.Blueprint) {
	logging.Info("BlueprintName: %v\n", bp.BlueprintName)
	logging.Info("GhpcVersion: %v\n", bp.GhpcVersion)
	logging.Info("Validators: %v\n", bp.Validators)
	logging.Info("ValidationLevel: %v\n", bp.ValidationLevel)
	logging.Info("Vars: %v\n", bp.Vars)
	logging.Info("Groups: %v\n", bp.Groups)
	logging.Info("TerraformBackendDefaults: %v\n", bp.TerraformBackendDefaults)
	logging.Info("TerraformProviders: %v\n", bp.TerraformProviders)
	logging.Info("ToolkitModulesURL: %v\n", bp.ToolkitModulesURL)
	logging.Info("ToolkitModulesVersion: %v\n", bp.ToolkitModulesVersion)
	logging.Info("YamlCtx: %v\n", bp.YamlCtx)
}
