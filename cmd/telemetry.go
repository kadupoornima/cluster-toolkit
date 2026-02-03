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

package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// internalTelemetryCmd represents the hidden telemetry uploader command. It is created to separate the telemetry call from the CLI call.
var internalTelemetryCmd = &cobra.Command{
	Use:    "internal-telemetry",
	Hidden: true, // Crucial: Hide from user help output
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		defer os.Remove(filePath) // Cleanup the payload file

		// 1. Read the upload configuration
		data, err := os.ReadFile(filePath)
		if err != nil {
			return
		}

		var config struct {
			URL     string            `json:"url"`
			Method  string            `json:"method"`
			Headers map[string]string `json:"headers"`
			Data    string            `json:"data"`
			Params  map[string]string `json:"params"`
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return
		}

		// 2. Prepare the Request
		req, err := http.NewRequest(config.Method, config.URL, strings.NewReader(config.Data))
		if err != nil {
			return
		}

		for k, v := range config.Headers {
			req.Header.Set(k, v)
		}

		q := req.URL.Query()
		for k, v := range config.Params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()

		// 3. Execute (Fire and Forget)
		client := &http.Client{}
		_, _ = client.Do(req)
	},
}

func init() {
	rootCmd.AddCommand(internalTelemetryCmd)
}
