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
	"encoding/json"
	"fmt"
	"hpc-toolkit/pkg/config"
	"net/http"
	"strings"
	"time"
)

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

func Flush() {
	if !config.IsTelemetryEnabled() {
		return
	}
	PrintLogRequest() // remove

	payload := ConstructPayload()

	jsonData, err := json.Marshal(payload)

	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	client := &http.Client{
		Timeout: HttpServerTimeout,
	}

	resp, reqErr := client.Post(HttpDummy, "application/json", strings.NewReader(string(jsonData)))

	if reqErr != nil {
		fmt.Printf("Request failed: %v\n", reqErr)
		return
	}
	resp.Body.Close()
}

func FlushOffline() {
	if config.IsTelemetryEnabled() {
		PrintLogRequest()
	}
}
