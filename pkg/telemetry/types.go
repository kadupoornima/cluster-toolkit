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

type LogRequest struct {
	ClientInfo    ClientInfo `json:"client_info"`
	LogSourceName string     `json:"log_source_name"`
	RequestTimeMs int64      `json:"request_time_ms"`
	LogEvent      []LogEvent `json:"log_event"`
}

type ClientInfo struct {
	ClientType string `json:"client_type"` // Value: "CLUSTER_TOOLKIT"
}

type LogEvent struct {
	EventTimeMs         int64  `json:"event_time_ms"`
	SourceExtensionJson string `json:"source_extension_json"` // Serialized ConcordEvent
}

// ConcordEvent matches the internal CloudMill proto structure
type ConcordEvent struct {
	ReleaseVersion  string          `json:"release_version"`
	ConsoleType     string          `json:"console_type"` // Value: "CLUSTER_TOOLKIT"
	ClientInstallId string          `json:"client_install_id"`
	EventType       string          `json:"event_type"` // e.g., "commands"
	EventName       string          `json:"event_name"` // e.g., "start", "complete"
	EventMetadata   []MetadataEntry `json:"event_metadata"`
}

type MetadataEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
