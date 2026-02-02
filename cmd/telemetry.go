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
	"bytes"
	"fmt"
	"hpc-toolkit/pkg/telemetry"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var telemetryEnabled bool = true

func init() {
	telemetryEnabled = true
}

func addTelemetryFlag(flagset *pflag.FlagSet) {
	flagset.BoolVar(&telemetryEnabled, "telemetry", true, "Enable usage telemetry.")
}

func initTelemetry(cmd *cobra.Command, args []string) {
	filePath := args[0]
	// Ensure the temp file is deleted regardless of success/failure
	defer os.Remove(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://play.googleapis.com/log", bytes.NewBuffer(data))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	// Match the User-Agent format used in XPK/gcloud
	req.Header.Set("User-Agent", "ClusterToolkit/1.0.0")

	// Add required query parameters
	q := req.URL.Query()
	q.Add("format", "json_proto")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err == nil {
		// Best practice: close body to prevent resource leaks
		resp.Body.Close()
	}
}

func addTelemetryEvent(eventName string, cmd *cobra.Command, args []string, version string) {
	if cmd.Flags().Lookup("telemetry").Value.String() == "true" {
		telemetry.AddEvent(eventName, map[string]string{
			"COMMAND":     cmd.CommandPath(),
			"CTK_VERSION": version,
			"ARGS":        fmt.Sprintf("%v", args),
			"EXIT_CODE":   "0", // fix
		})
		telemetry.Flush()
	}
}
