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

// The following implementation is done for sending one LogEvent per LogRequest as per the telemetry logic.

package telemetry

import (
	"encoding/json"
	"hpc-toolkit/pkg/config"
	"time"

	"github.com/spf13/cobra"

	"github.com/google/uuid"
)

type ClientInfo struct {
	client_type string
}

type EventMetadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LogEvent struct {
	EventTimeMs         int64  `json:"event_time_ms"`
	SourceExtensionJson string `json:"source_extension_json"` // Contains event metadata as key-value pairs.
}

type LogRequest struct {
	RequestTimeMs int64      `json:"request_time_ms"`
	ClientInfo    ClientInfo `json:"client_info"`
	LogSourceName string     `json:"log_source_name"`
	LogEvents     []LogEvent `json:"log_events"`
}

var (
	// logRequest     LogRequest
	logEvent       LogEvent
	eventMetadata  []EventMetadata = make([]EventMetadata, 0)
	eventStartTime time.Time
	eventEndTime   time.Time
)

func CollectPreMetrics(cmd *cobra.Command, args []string) {
	logEvent.EventTimeMs = time.Now().UnixMilli()
	eventStartTime = time.Now()

	eventMetadata = append(eventMetadata, []EventMetadata{
		{Key: "CLUSTER_TOOLKIT_COMMAND_NAME", Value: cmd.Name()},
		{Key: "CLUSTER_TOOLKIT_COMMAND_LINE_ARGS", Value: "testCommandLineArgs"},
		{Key: "CLUSTER_TOOLKIT_BLUEPRINT", Value: getBlueprintName()},
	}...)
}

func getBlueprintName() string {
	return "testBlueprintName"
}

// func calculateLatency() int64 {
// 	return 0
// }

func CollectPostMetrics(errorCode int) {

	eventMetadata = append(eventMetadata, []EventMetadata{
		{Key: "CLUSTER_TOOLKIT_EXIT_CODE", Value: string(rune(errorCode))},
		{Key: "CLUSTER_TOOLKIT_LATENCY", Value: string(rune(eventEndTime.Sub(eventStartTime).Milliseconds()))},
		{Key: "CLUSTER_TOOLKIT_EXECUTION_TIME", Value: string(rune(eventEndTime.UnixNano()))},
	}...)
	// exitCode = errorCode
}

func ConstructPayload() LogRequest {
	eventMetadata = append(eventMetadata, []EventMetadata{
		{Key: "CLUSTER_TOOLKIT_SESSION_ID", Value: uuid.New().String()},
		{Key: "CLUSTER_TOOLKIT_CLIENT_ID", Value: ensureClientId()},
		{Key: "CLUSTER_TOOLKIT_VERSION", Value: config.GetToolkitVersion()},
		// {Key: "CLUSTER_TOOLKIT_EXIT_CODE", Value: string(rune(exitCode))},

		// 	// The command used by the user.
		// 	CLUSTER_TOOLKIT_COMMAND_NAME = 4;

		// 	// The command line args sent by the user.
		// 	CLUSTER_TOOLKIT_COMMAND_LINE_ARGS = 5;

		// 	// Standard blueprint names taken as is, others hashed or categorized as
		// 	// Custom (No PII).
		// 	CLUSTER_TOOLKIT_BLUEPRINT = 6;

		// 	// Scheduler Type (e.g. GKE, Slurm).
		// 	CLUSTER_TOOLKIT_SCHEDULER = 7;

		// 	// Primary machine type
		// 	CLUSTER_TOOLKIT_MACHINE_TYPE = 8;

		// 	// The provisioning mode (consumption model) used by the user.
		// 	CLUSTER_TOOLKIT_PROVISIONING_MODE = 9;

		// 	// The exit code of the command.
		// 	CLUSTER_TOOLKIT_EXIT_CODE = 10;

		// 	// The latency of the event in milliseconds.
		// 	CLUSTER_TOOLKIT_LATENCY = 11;

		// 	// Time of the execution.
		// 	CLUSTER_TOOLKIT_EXECUTION_TIME = 12;

		// 	// OS name
		// 	CLUSTER_TOOLKIT_OS_NAME = 13;

		// 	// OS version
		// 	CLUSTER_TOOLKIT_OS_VERSION = 14;
		//   }

	}...)

	sourceExtensionJSON, err := json.Marshal(map[string]interface{}{
		"event_type":     "GCluster CLI",
		"event_name":     "GCluster CLI command",
		"event_metadata": eventMetadata,
	})
	if err != nil {
		// Handle error
		return LogRequest{}
	}

	logEvent := LogEvent{
		EventTimeMs:         time.Now().UnixMilli(),
		SourceExtensionJson: string(sourceExtensionJSON),
	}

	return LogRequest{
		RequestTimeMs: time.Now().UnixMilli(),
		ClientInfo:    ClientInfo{client_type: "CLUSTER_TOOLKIT"},
		LogSourceName: "CONCORD",
		LogEvents:     []LogEvent{logEvent},
	}

}

// return json.dumps({
// 	"client_info": {"client_type": "CLUSTER_TOOLKIT"},
// 	"log_source_name": "CONCORD",
// 	"request_time_ms": int(time.time() * 1000),
// 	"log_event": serialized_events,
// })

// "event_time_ms": int(event.time * 1000),
//         "source_extension_json": json.dumps({
//             **base_concord_event,

// "release_version": xpk_version,
// "console_type": "XPK",
// "client_install_id": _ensure_client_id(),

//             "event_type": event.type,
//             "event_name": event.name,
//             "event_metadata": [
//                 {"key": key.value, "value": value}
//                 for key, value in metadata.items()
//             ],
//         }),

func ensureClientId() string {
	if config.GetClientId() != "" {
		return config.GetClientId()
	}
	return config.SetClientId()
}
