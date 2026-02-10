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

import "time"

const (
	ClearcutProdURL       = "https://play.googleapis.com/log"
	ClearcutStagingURL    = "https://play.googleapis.com/staging/log"
	ClearcutAltProdURL    = "https://play.google.com/log?format=json&hasfast=true"
	ClearcutAltStagingURL = "https://play.google.com/staging/log?format=json&hasfast=true"
	ClearcutLocalURL      = "http://localhost:27910/log"
	HttpDummy             = "http://127.0.0.1:8888"
	LogSourceEnum         = 113
	ClientType            = "CLUSTER_TOOLKIT"
	configDirName         = "cluster-toolkit"
	HttpServerTimeout     = 10 * time.Second
)

type ClientInfo struct {
	ClientType string `json:"client_type"`
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
