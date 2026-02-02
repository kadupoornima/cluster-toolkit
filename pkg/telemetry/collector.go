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
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
)

var (
	sessionID = uuid.New().String()
	events    []LogEvent
)

func AddEvent(name string, metadata map[string]string) {
	// 1. Check if telemetry is enabled in config
	if metadata["telemetry"] == "false" {
		return
	}

	// 2. Prepare Metadata
	baseMeta := []MetadataEntry{
		{"SESSION_ID", sessionID},
		{"OS", runtime.GOOS},
	}

	for k, v := range metadata {
		baseMeta = append(baseMeta, MetadataEntry{Key: k, Value: v})
	}

	// 3. Create Concord Event
	concordEvent := ConcordEvent{
		ConsoleType:     "CLUSTER_TOOLKIT",
		ClientInstallId: "GET_FROM_CONFIG", // Persistent Client ID
		EventType:       "commands",
		EventName:       name,
		EventMetadata:   baseMeta,
	}

	concordJson, _ := json.Marshal(concordEvent)

	// 4. Append to local buffer
	events = append(events, LogEvent{
		EventTimeMs:         time.Now().UnixMilli(),
		SourceExtensionJson: string(concordJson),
	})
}

// Flush writes the events to a temp file and triggers the background uploader
func Flush() {
	if len(events) == 0 {
		return
	}

	// 1. Construct final payload
	payload := LogRequest{
		ClientInfo:    ClientInfo{ClientType: "CLUSTER_TOOLKIT"},
		LogSourceName: "CONCORD",
		RequestTimeMs: time.Now().UnixMilli(),
		LogEvent:      events,
	}

	// 2. Write to Temp File to pass payload to the detached process
	f, _ := os.CreateTemp("", "gcluster-telemetry-*.json")
	defer f.Close()
	json.NewEncoder(f).Encode(payload)

	// 3. Trigger Background Upload
	triggerBackgroundUpload(f.Name())
}
