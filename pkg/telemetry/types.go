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
	LogEvent      []LogEvent `json:"log_event"`
}

const (
	USER_ID              = "CLUSTER_TOOLKIT_USER_ID"
	COMMAND_NAME         = "CLUSTER_TOOLKIT_COMMAND_NAME"
	COMMAND_FLAGS        = "CLUSTER_TOOLKIT_COMMAND_FLAGS"
	BLUEPRINT            = "CLUSTER_TOOLKIT_BLUEPRINT"
	DEPLOYMENT_FILE      = "CLUSTER_TOOLKIT_DEPLOYMENT_FILE"
	BILLING_ACCOUNT      = "CLUSTER_TOOLKIT_BILLING_ACCOUNT"
	IS_GKE               = "CLUSTER_TOOLKIT_IS_GKE"
	IS_SLURM             = "CLUSTER_TOOLKIT_IS_SLURM"
	IS_VM_INSTANCE       = "CLUSTER_TOOLKIT_IS_VM_INSTANCE"
	MACHINE_TYPE         = "CLUSTER_TOOLKIT_MACHINE_TYPE"
	REGION               = "CLUSTER_TOOLKIT_REGION"
	ZONE                 = "CLUSTER_TOOLKIT_ZONE"
	PROVISIONING_MODE    = "CLUSTER_TOOLKIT_PROVISIONING_MODE"
	MODULES              = "CLUSTER_TOOLKIT_MODULES"
	OS_NAME              = "CLUSTER_TOOLKIT_OS_NAME"
	OS_VERSION           = "CLUSTER_TOOLKIT_OS_VERSION"
	TERRAFORM_VERSION    = "CLUSTER_TOOLKIT_TERRAFORM_VERSION"
	IS_INTERNAL_USER     = "CLUSTER_TOOLKIT_IS_INTERNAL_USER"
	DEPLOYED_FROM_SOURCE = "CLUSTER_TOOLKIT_DEPLOYED_FROM_SOURCE"
	DEPLOYED_FROM_BINARY = "CLUSTER_TOOLKIT_DEPLOYED_FROM_BINARY"
	IS_TEST_DATA         = "CLUSTER_TOOLKIT_IS_TEST_DATA"
	RUNTIME_MS           = "CLUSTER_TOOLKIT_RUNTIME_MS"
	EXIT_CODE            = "CLUSTER_TOOLKIT_EXIT_CODE"
)
