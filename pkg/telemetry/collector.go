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
	crm "google.golang.org/api/cloudresourcemanager/v1"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	metadata                   = make(map[string]string)
	blueprint                  config.Blueprint
	modulesList                []string
	eventStartTime             time.Time
	IsGkeModulePatterns        = []string{".*gke-node-pool.*", ".*gke-cluster.*"}
	IsSlurmModulePatterns      = []string{".*schedmd-slurm-gcp-.*"}
	IsVmInstanceModulePatterns = []string{".*compute/vm-instance.*"}
	MachineTypeModulePatterns  = []string{".*gke-node-pool.*", ".*schedmd-slurm-gcp-.*-nodeset.*", ".*compute/vm-instance.*"} // module patterns where machine_type can be specified
)

// googleOrgID is the canonical Google.com organization ID
const googleOrgID = "433637338589"

func CollectPreMetrics(cmd *cobra.Command, args []string) {
	eventStartTime = time.Now()
	blueprint = getBlueprint(args)
	modulesList = getModulesList(blueprint)
	logBlueprint(blueprint)

	metadata[USER_ID] = getUserId()
	metadata[COMMAND_NAME] = getCommandName(cmd)
	metadata[COMMAND_FLAGS] = getCmdFlags(cmd)
	metadata[BLUEPRINT] = getBlueprintName()
	metadata[DEPLOYMENT_FILE] = getDeploymentFile()
	metadata[IS_GKE] = getIsGke()
	metadata[IS_SLURM] = getIsSlurm()
	metadata[IS_VM_INSTANCE] = getIsVmInstance()
	metadata[MACHINE_TYPE] = getMachineType()
	metadata[REGION] = getRegion()
	metadata[ZONE] = getZone()
	metadata[PROVISIONING_MODE] = getProvisioningMode()
	metadata[MODULES] = getModules()
	metadata[OS_NAME] = getOSName()
	metadata[OS_VERSION] = getOSVersion()
	metadata[TERRAFORM_VERSION] = getTerraformVersion()
	metadata[DEPLOYED_FROM_SOURCE] = getDeployedFromSource()
	metadata[DEPLOYED_FROM_BINARY] = getDeployedFromBinary()
	metadata[IS_TEST_DATA] = getIsTestData()

}

func CollectPostMetrics(errorCode int) {
	metadata[RUNTIME_MS] = getRuntime()
	metadata[EXIT_CODE] = strconv.Itoa(errorCode)
	// Collecting these metrics after the run to ensure the API calls do not add additional latency.
	metadata[BILLING_ACCOUNT] = getBillingAccount()
	metadata[IS_INTERNAL_USER] = getIsInternalUser()
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

func getBlueprintName() string {
	return blueprint.BlueprintName
}

func getDeploymentFile() string {

	return "test"
}

func getIsGke() string {
	isGke := false
	isGke = slices.ContainsFunc(modulesList, func(s string) bool {
		for _, pattern := range IsGkeModulePatterns {
			match, _ := regexp.MatchString(pattern, s)
			isGke = isGke || match
		}
		return isGke
	})
	return strconv.FormatBool(isGke)
}

func getIsSlurm() string {
	isSlurm := false
	isSlurm = slices.ContainsFunc(modulesList, func(s string) bool {
		for _, pattern := range IsSlurmModulePatterns {
			match, _ := regexp.MatchString(pattern, s)
			isSlurm = isSlurm || match
		}
		return isSlurm
	})
	return strconv.FormatBool(isSlurm)
}

func getIsVmInstance() string {
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

func getMachineType() string {
	machine_types := make([]string, 0)
	modules := getModulesFromPatterns(MachineTypeModulePatterns)
	for _, m := range modules {
		if m.Settings.Has("machine_type") {
			machine_types = append(machine_types, m.Settings.Get("machine_type").AsString())
		}
	}
	return strings.Join(machine_types, ",")
}

func getRegion() string {
	if blueprint.Vars.Has("region") {
		return blueprint.Vars.Get("region").AsString()
	}
	return ""
}

func getZone() string {
	if blueprint.Vars.Has("zone") {
		return blueprint.Vars.Get("zone").AsString()
	}
	return ""
}

func getProvisioningMode() string {

	return "test"
}

func getModules() string {
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

	return strings.Join(getModulesList(blueprint), ",")
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

func getTerraformVersion() string {
	// tfProviders := blueprint.TerraformProviders

	return strconv.Itoa(len(blueprint.TerraformProviders))
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

func getBillingAccount() string {
	projectID := getProjectId(blueprint)
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
}

// getIsInternalUser returns "true" if the GCP project belongs to the Google.com organization.
func getIsInternalUser() string {
	projectID := getProjectId(blueprint)
	if projectID == "" {
		return "false"
	}
	ctx := context.Background()
	service, err := crm.NewService(ctx)
	if err != nil {
		return "false"
	}

	// Fetch the ancestry of the project
	req := &crm.GetAncestryRequest{}
	resp, err := service.Projects.GetAncestry(projectID, req).Do()
	if err != nil {
		// This can fail if the user lacks IAM permissions or the project doesn't exist
		return "false"
	}

	// Traverse the ancestors from bottom (the project) to top (the organization)
	for _, ancestor := range resp.Ancestor {
		if ancestor.ResourceId.Type == "organization" && ancestor.ResourceId.Id == googleOrgID {
			return "true"
		}
	}

	return "false"
}

func getRuntime() string {
	eventEndTime := time.Now()
	return strconv.FormatInt(eventEndTime.Sub(eventStartTime).Milliseconds(), 10)
}

/****************************************************************************************************/
/************************************** Utility functions *******************************************/
/****************************************************************************************************/

func getBlueprint(args []string) config.Blueprint {
	bp, _, _ := config.NewBlueprint(args[0])
	return bp
}

func getModulesFromPatterns(patterns []string) []config.Module {
	modules := make([]config.Module, 0)
	for _, m := range config.GetAllModules(&blueprint) {
		for _, p := range patterns {
			matched, _ := regexp.Match(p, []byte(m.Source))
			if matched {
				modules = append(modules, m)
			}
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

func getProjectId(bp config.Blueprint) string {
	if bp.Vars.Has("project_id") {
		return bp.Vars.Get("project_id").AsString()
	}
	return ""
}

// getProjectBillingAccount fetches the billing account associated with a given GCP project in the format "billingAccounts/{billing_account_id}". If billing is disabled for the project, this will return an empty string.
func getProjectBillingAccount(ctx context.Context, projectID string) (string, error) {
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
