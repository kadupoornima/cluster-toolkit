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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	metadata       = make(map[string]string)
	bp             config.Blueprint
	eventStartTime time.Time
)

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
func CollectPreMetrics(cmd *cobra.Command, args []string) {
	eventStartTime = time.Now()
	bp = getBlueprint(args)
	logBlueprint(bp)

	metadata["CLUSTER_TOOLKIT_USER_ID"] = getUserId()
	metadata["CLUSTER_TOOLKIT_COMMAND_NAME"] = getCommandName(cmd)
	metadata["CLUSTER_TOOLKIT_COMMAND_FLAGS"] = getCmdFlags(cmd)
	metadata["CLUSTER_TOOLKIT_BLUEPRINT"] = getBlueprintName(bp)
	metadata["CLUSTER_TOOLKIT_DEPLOYMENT_FILE"] = getDeploymentFile()
	metadata["CLUSTER_TOOLKIT_IS_GKE"] = getIsGke(bp)
	metadata["CLUSTER_TOOLKIT_IS_SLURM"] = getIsSlurm()
	metadata["CLUSTER_TOOLKIT_IS_VM_INSTANCE"] = getIsVmInstance()
	metadata["CLUSTER_TOOLKIT_MACHINE_TYPE"] = getMachineType()
	metadata["CLUSTER_TOOLKIT_REGION"] = getRegion(bp)
	metadata["CLUSTER_TOOLKIT_ZONE"] = getZone(bp)
	metadata["CLUSTER_TOOLKIT_PROVISIONING_MODE"] = getProvisioningMode()
	metadata["CLUSTER_TOOLKIT_MODULES"] = getModules(bp)
	metadata["CLUSTER_TOOLKIT_OS_NAME"] = getOSName()
	metadata["CLUSTER_TOOLKIT_OS_VERSION"] = getOSVersion()
	metadata["CLUSTER_TOOLKIT_TERRAFORM_VERSION"] = getTerraformVersion(bp)
	metadata["CLUSTER_TOOLKIT_IS_INTERNAL_USER"] = getIsInternalUser()
	metadata["CLUSTER_TOOLKIT_DEPLOYED_FROM_SOURCE"] = getDeployedFromSource()
	metadata["CLUSTER_TOOLKIT_DEPLOYED_FROM_BINARY"] = getDeployedFromBinary()
	metadata["CLUSTER_TOOLKIT_IS_TEST_DATA"] = getIsTestData()

}

func CollectPostMetrics(errorCode int) {
	metadata["CLUSTER_TOOLKIT_RUNTIME_MS"] = getRuntime()
	metadata["CLUSTER_TOOLKIT_EXIT_CODE"] = strconv.Itoa(errorCode)
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
	GkeModulePatterns   = []string{".*gke-node-pool.*", ".*gke-cluster.*"}
	SlurmModulePatterns = []string{".*schedmd-slurm-gcp-.*"}
)

func getIsGke(bp config.Blueprint) string {
	modules := getModulesList(bp)
	isGke := false
	isGke = slices.ContainsFunc(modules, func(s string) bool {
		for _, pattern := range GkeModulePatterns {
			match, _ := regexp.MatchString(pattern, s)
			isGke = isGke || match
		}
		return isGke
	})
	return strconv.FormatBool(isGke)
}

func getIsSlurm() string {
	modules := getModulesList(bp)
	isSlurm := false
	isSlurm = slices.ContainsFunc(modules, func(s string) bool {
		for _, pattern := range SlurmModulePatterns {
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
func getMachineType() string {

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

func getSchedulers() string {
	schedulers := make([]string, 0)
	schedulers = append(schedulers, "testSchedulers1")
	schedulers = append(schedulers, "testScheduler2")
	return strings.Join(schedulers, ",")
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
