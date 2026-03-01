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
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	metadata = make(map[string]string)
	bp       config.Blueprint
)

func CollectPreMetrics(cmd *cobra.Command, args []string) {
	bp = getBlueprint(args)

	metadata["CLUSTER_TOOLKIT_EVENT_ID"] = getEventId()
	metadata["CLUSTER_TOOLKIT_USER_ID"] = getUserId()
	metadata["CLUSTER_TOOLKIT_VERSION"] = getToolkitVersion()
	metadata["CLUSTER_TOOLKIT_COMMAND_NAME"] = getCommandName(cmd)
	metadata["CLUSTER_TOOLKIT_COMMAND_FLAGS"] = getCmdFlags(cmd)
	metadata["CLUSTER_TOOLKIT_BLUEPRINT"] = getBlueprintName(bp)
	metadata["CLUSTER_TOOLKIT_DEPLOYMENT_FILE"] = getDeploymentFile(args)
	metadata["CLUSTER_TOOLKIT_IS_GKE"] = getIsGke(args)
	metadata["CLUSTER_TOOLKIT_IS_SLURM"] = getIsSlurm(args)
	metadata["CLUSTER_TOOLKIT_IS_VM_INSTANCE"] = getIsVmInstance(args)
	metadata["CLUSTER_TOOLKIT_MACHINE_TYPE"] = getMachineType(args)
	metadata["CLUSTER_TOOLKIT_REGION"] = getRegion(args)
	metadata["CLUSTER_TOOLKIT_ZONE"] = getZone(args)
	metadata["CLUSTER_TOOLKIT_PROVISIONING_MODE"] = getProvisioningMode(args)
	metadata["CLUSTER_TOOLKIT_MODULES"] = getModules(bp)
	metadata["CLUSTER_TOOLKIT_OS_NAME"] = getOSName()
	metadata["CLUSTER_TOOLKIT_OS_VERSION"] = getOSVersion()
	metadata["CLUSTER_TOOLKIT_TERRAFORM_VERSION"] = getTerraformVersion(args)
	metadata["CLUSTER_TOOLKIT_IS_INTERNAL_USER"] = getIsInternalUser(args)
	metadata["CLUSTER_TOOLKIT_DEPLOYED_FROM_SOURCE"] = getDeployedFromSource(args)
	metadata["CLUSTER_TOOLKIT_DEPLOYED_FROM_BINARY"] = getDeployedFromBinary(args)
	metadata["CLUSTER_TOOLKIT_IS_TEST_DATA"] = getIsTestData()

}

func CollectPostMetrics(errorCode int) {
	metadata["CLUSTER_TOOLKIT_RUNTIME_MS"] = getRuntime()
	metadata["CLUSTER_TOOLKIT_EXIT_CODE"] = strconv.Itoa(errorCode)
}

func getEventId() string {
	return uuid.New().String()
}

func getCommandName(cmd *cobra.Command) string {
	return cmd.Name()
}

func getBlueprint(args []string) config.Blueprint {
	bp, _, _ := config.NewBlueprint(args[0])
	return bp
}
func getDeploymentFile(args []string) string {

	return args[0]
}
func getIsGke(args []string) string {
	return args[0]
}
func getIsSlurm(args []string) string {

	return args[0]
}
func getIsVmInstance(args []string) string {

	return args[0]
}
func getMachineType(args []string) string {

	return args[0]
}
func getRegion(args []string) string {

	return args[0]
}
func getZone(args []string) string {

	return args[0]
}
func getProvisioningMode(args []string) string {

	return args[0]
}
func getIsInternalUser(args []string) string {

	return args[0]
}
func getTerraformVersion(args []string) string {
	return args[0]
}

func getDeployedFromSource(args []string) string {
	return args[0]
}
func getDeployedFromBinary(args []string) string {
	return args[0]
}
func getIsTestData() string {
	return "true"
}

func getModules(bp config.Blueprint) string {
	moduleInfos := make([]config.Module, 0)
	modules := make([]string, 0)
	moduleInfos = append(moduleInfos, config.GetAllModules(&bp)...)
	for _, module := range moduleInfos {
		modules = append(modules, string(module.Source))
	}
	return strings.Join(modules, ",")
}

func getToolkitVersion() string {
	return config.GetToolkitVersion()
}

func getCmdFlags(cmd *cobra.Command) string {
	flags := make([]string, 0)
	cmd.LocalFlags().Visit(func(f *pflag.Flag) {
		flags = append(flags, f.Name)
	})
	return strings.Join(flags, ",")
}

func getUserId() string {
	return config.GetPersistentUserId()
}

func getBlueprintName(bp config.Blueprint) string {
	return bp.BlueprintName
}

func getRuntime() string {
	eventEnd := time.Now()
	eventStart, _ := time.Parse(time.RFC3339, metadata["CLUSTER_TOOLKIT_EVENT_TIME"])

	return strconv.FormatInt(eventEnd.Sub(eventStart).Milliseconds(), 10)
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
