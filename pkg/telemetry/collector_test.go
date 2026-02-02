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
	"runtime"
	"testing"

	"github.com/google/uuid"
)

func TestAddEvent(t *testing.T) {
	// Save the original sessionID and events, and restore them after the test.
	originalSessionID := sessionID
	originalEvents := events
	defer func() {
		sessionID = originalSessionID
		events = originalEvents
	}()

	// Generate a new UUID for testing.
	testSessionID := uuid.New().String()
	sessionID = testSessionID

	// Initialize an empty slice for events.
	events = []LogEvent{}

	// Define test cases.
	testCases := []struct {
		name           string
		metadata       map[string]string
		expectEvent    bool
		expectMetadata map[string]string
	}{
		{
			name: "Telemetry Enabled",
			metadata: map[string]string{
				"key1":      "value1",
				"telemetry": "true", // or absent
			},
			expectEvent: true,
			expectMetadata: map[string]string{
				"key1":       "value1",
				"SESSION_ID": testSessionID,
				"OS":         runtime.GOOS,
			},
		},
		{
			name: "Telemetry Disabled",
			metadata: map[string]string{
				"key1":      "value1",
				"telemetry": "false",
			},
			expectEvent: false,
		},
		{
			name:        "No Metadata",
			metadata:    map[string]string{},
			expectEvent: true,
			expectMetadata: map[string]string{
				"SESSION_ID": testSessionID,
				"OS":         runtime.GOOS,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			AddEvent("testEvent", tc.metadata)

			if tc.expectEvent {
				if len(events) != 1 {
					t.Errorf("Expected 1 event, got %d", len(events))
					return
				}

				var concordEvent ConcordEvent
				err := json.Unmarshal([]byte(events[0].SourceExtensionJson), &concordEvent)
				if err != nil {
					t.Fatalf("Failed to unmarshal ConcordEvent: %v", err)
				}

				// Verify the event name.
				if concordEvent.EventName != "testEvent" {
					t.Errorf("Event name mismatch: expected 'testEvent', got '%s'", concordEvent.EventName)
				}

				// Verify metadata.
				for k, v := range tc.expectMetadata {
					found := false
					for _, meta := range concordEvent.EventMetadata {
						if meta.Key == k {
							if meta.Value != v {
								t.Errorf("Metadata mismatch for key '%s': expected '%s', got '%s'", k, v, meta.Value)
							}
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Metadata key '%s' not found", k)
					}
				}
			} else {
				if len(events) != 0 {
					t.Errorf("Expected no events, got %d", len(events))
				}
			}

			// Reset events for the next test case.
			events = []LogEvent{}
		})
	}
}
