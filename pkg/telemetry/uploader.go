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
)

var httpConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Data    string            `json:"data"`
	Params  map[string]string `json:"params"`
}

func Flush() {
	if !config.IsTelemetryEnabled() || logRequest.LogEvents == nil {
		return
	}
	PrintLogRequest()

	payload := ConstructPayload()

	if payload.LogEvents == nil {
		return
	}

	// 2. Marshall the struct to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	uploadConfig := map[string]interface{}{
		"url":     HttpDummy,
		"method":  "POST",
		"headers": map[string]string{"User-Agent": "gcluster-telemetry/1.0"},
		"data":    jsonData,
		"params":  map[string]string{"format": "json_proto"},
	}

	if err := json.Unmarshal(jsonData, &uploadConfig); err != nil {
		return
	}

	// 2. Prepare the Request
	req, err := http.NewRequest(httpConfig.Method, httpConfig.URL, strings.NewReader(httpConfig.Data))
	if err != nil {
		return
	}

	for k, v := range httpConfig.Headers {
		req.Header.Set(k, v)
	}

	q := req.URL.Query()
	for k, v := range httpConfig.Params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// 3. Create the request with a timeout (best practice)
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	_, reqErr := client.Do(req)

	if reqErr != nil {
		fmt.Printf("Request failed: %v\n", reqErr)
		return
	}
}

func FlushOffline() {
	if config.IsTelemetryEnabled() {
		PrintLogRequest()
	}
}
