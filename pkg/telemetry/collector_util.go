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
	"fmt"

	"hpc-toolkit/pkg/config"
	"hpc-toolkit/pkg/logging"
	"os"
	"os/exec"
	"regexp"
	"strings"

	billing "cloud.google.com/go/billing/apiv1"
	"cloud.google.com/go/billing/apiv1/billingpb"
)

func getBlueprint(args []string) config.Blueprint {
	bp, _, _ := config.NewBlueprint(args[0])
	return bp
}

func getEventMetadataKVPairs(sourceMetadata map[string]string) []map[string]string {
	eventMetadata := make([]map[string]string, 0)
	for k, v := range sourceMetadata {
		eventMetadata = append(eventMetadata, map[string]string{
			"key":   k,
			"value": v,
		})
	}
	return eventMetadata
}

func getModulesWithPattern(pattern string, bp config.Blueprint) []config.Module {
	modules := make([]config.Module, 0)
	for _, m := range config.GetAllModules(&bp) {
		matched, _ := regexp.Match(pattern, []byte(m.Source))
		if matched {
			logging.Info("Source: %v", m.Source)
			logging.Info("Items: %v", m.Settings.Items())
			// logging.Info("%v", m.Settings.Get("machine_type").GetAttr())
			logging.Info("Keys: %v", m.Settings.Keys())
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

func getProjectId(bp config.Blueprint) string {
	logging.Info("Keys:\n%v", bp.Vars.Keys())
	logging.Info("Items:\n%v", bp.Vars.Items())
	if bp.Vars.Has("project_id") {
		return bp.Vars.Get("project_id").AsString()
	}
	return ""
}

// getProjectBillingAccount fetches the billing account associated with a given GCP project in the format "billingAccounts/{billing_account_id}". If billing is disabled for the project, this will return an empty string.
func getProjectBillingAccount(ctx context.Context, projectID string) (string, error) {
	logging.Info("1-HIIII")
	client, err := billing.NewCloudBillingClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create billing client: %w", err)
	}
	defer client.Close()
	logging.Info("2-HIIII")
	req := &billingpb.GetProjectBillingInfoRequest{
		Name: fmt.Sprintf("projects/%s", projectID),
	}
	logging.Info("3-HIIII\n%v", req)
	info, err := client.GetProjectBillingInfo(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get billing info for project %s: %w", projectID, err)
	}
	logging.Info("4-HIIII")
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

// func logBlueprint(bp config.Blueprint) {
// 	logging.Info("BlueprintName: %v\n", bp.BlueprintName)
// 	logging.Info("GhpcVersion: %v\n", bp.GhpcVersion)
// 	logging.Info("Validators: %v\n", bp.Validators)
// 	logging.Info("ValidationLevel: %v\n", bp.ValidationLevel)
// 	logging.Info("Vars: %v\n", bp.Vars)
// 	logging.Info("Groups: %v\n", bp.Groups)
// 	logging.Info("TerraformBackendDefaults: %v\n", bp.TerraformBackendDefaults)
// 	logging.Info("TerraformProviders: %v\n", bp.TerraformProviders)
// 	logging.Info("ToolkitModulesURL: %v\n", bp.ToolkitModulesURL)
// 	logging.Info("ToolkitModulesVersion: %v\n", bp.ToolkitModulesVersion)
// 	logging.Info("YamlCtx: %v\n", bp.YamlCtx)
// }
