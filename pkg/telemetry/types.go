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

type EventMetadata struct {
	CLUSTER_TOOLKIT_SESSION_ID        string `json:"cluster_toolkit_session_id"`
	CLUSTER_TOOLKIT_CLIENT_ID         string `json:"cluster_toolkit_client_id"`
	CLUSTER_TOOLKIT_VERSION           string `json:"cluster_toolkit_version"`
	CLUSTER_TOOLKIT_COMMAND_NAME      string `json:"cluster_toolkit_command_name"`
	CLUSTER_TOOLKIT_COMMAND_LINE_ARGS string `json:"cluster_toolkit_command_line_args"`
	CLUSTER_TOOLKIT_BLUEPRINT         string `json:"cluster_toolkit_blueprint"`
	CLUSTER_TOOLKIT_SCHEDULER         string `json:"cluster_toolkit_scheduler"`
	CLUSTER_TOOLKIT_MACHINE_TYPE      string `json:"cluster_toolkit_machine_type"`
	CLUSTER_TOOLKIT_PROVISIONING_MODE string `json:"cluster_toolkit_provisioning_mode"`
	CLUSTER_TOOLKIT_EXIT_CODE         int    `json:"cluster_toolkit_exit_code"`
	CLUSTER_TOOLKIT_LATENCY_MS        int64  `json:"cluster_toolkit_latency_ms"` // runtime
	CLUSTER_TOOLKIT_EXECUTION_TIME    string `json:"cluster_toolkit_execution_time"`
	CLUSTER_TOOLKIT_OS_NAME           string `json:"cluster_toolkit_os_name"`
	CLUSTER_TOOLKIT_OS_VERSION        string `json:"cluster_toolkit_os_version"`
	// add ? for datetime / time of request
	// add note in ticket that this is taken from g3 cluster_toolkit_key.proto. Should keep in sync.
}

type ClientInfo struct {
	client_type string
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
