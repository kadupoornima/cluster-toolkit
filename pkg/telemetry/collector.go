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
	"hpc-toolkit/pkg/config"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	metadata = make(map[string]string)
)

func CollectPreMetrics(cmd *cobra.Command, args []string) {
	metadata["CLUSTER_TOOLKIT_COMMAND_NAME"] = getCommandName(cmd)
	metadata["CLUSTER_TOOLKIT_COMMAND_LINE_ARGS"] = getCommandLineArgs(args)
	metadata["CLUSTER_TOOLKIT_SESSION_ID"] = getSessionId()
	metadata["CLUSTER_TOOLKIT_CLIENT_ID"] = getClientId()
	metadata["CLUSTER_TOOLKIT_VERSION"] = config.GetToolkitVersion()
	metadata["CLUSTER_TOOLKIT_BLUEPRINT"] = getBlueprintName()
	metadata["CLUSTER_TOOLKIT_DATE_TIME"] = time.Now().Format(time.RFC3339)
	metadata["CLUSTER_TOOLKIT_SCHEDULER"] = getSchedulers()
	metadata["CLUSTER_TOOLKIT_OS_NAME"] = getOSName()
	metadata["CLUSTER_TOOLKIT_OS_VERSION"] = getOSVersion()
}

func CollectPostMetrics(errorCode int) {
	metadata["CLUSTER_TOOLKIT_RUNTIME_MS"] = strconv.FormatInt(calculateRuntime(), 10)
	metadata["CLUSTER_TOOLKIT_EXIT_CODE"] = strconv.Itoa(errorCode)
}

func getCommandName(cmd *cobra.Command) string {
	return cmd.Name()
}

func getCommandLineArgs(args []string) string {
	return args[0]
}

func getSessionId() string {
	return uuid.New().String()
}

func getClientId() string {
	clientID, _ := config.GetPersistentClientID()
	return clientID
}

func getBlueprintName() string {
	return "testBlueprintName"
}

func calculateRuntime() int64 {
	eventEnd := time.Now()
	eventStart, _ := time.Parse(time.RFC3339, metadata["CLUSTER_TOOLKIT_EXECUTION_TIME"])

	return int64(eventEnd.Sub(eventStart).Milliseconds())
}

func getOSName() string {
	// return config.GetOSName()
	return "testOSName"
}

func getOSVersion() string {
	// return config.GetOSVersion()
	return "testOSVersion"
}

func getSchedulers() string {
	schedulers := make([]string, 0)
	schedulers = append(schedulers, "testSchedulers1")
	schedulers = append(schedulers, "testScheduler2")
	return strings.Join(schedulers, ",")
}
