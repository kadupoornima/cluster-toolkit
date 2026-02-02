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
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInit(t *testing.T) {
	collector = &MetricsCollector{
		events:    make([]*MetricsEvent, 0),
		sessionID: uuid.New().String(),
		clientID:  getOrGenerateClientID(),
		enabled:   true,
	}
	Init(false)
	if collector.enabled {
		t.Errorf("Init(false) failed, expected enabled to be false, got true")
	}

	Init(true)
	if !collector.enabled {
		t.Errorf("Init(true) failed, expected enabled to be true, got false")
	}
}

func TestLogStart(t *testing.T) {
	collector = &MetricsCollector{
		mu:        sync.Mutex{},
		enabled:   true,
		events:    make([]*MetricsEvent, 0),
		sessionID: uuid.New().String(),
		clientID:  getOrGenerateClientID(),
	}
	command := "test-command"
	LogStart(command)

	if len(collector.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(collector.events))
	}

	event := collector.events[0]
	if event.Type != "commands" {
		t.Errorf("Expected event type 'commands', got '%s'", event.Type)
	}
	if event.Name != "start" {
		t.Errorf("Expected event name 'start', got '%s'", event.Name)
	}
	if event.Metadata["command"] != command {
		t.Errorf("Expected command '%s', got '%s'", command, event.Metadata["command"])
	}
}

func TestLogComplete(t *testing.T) {
	collector = &MetricsCollector{
		mu:        sync.Mutex{},
		enabled:   true,
		events:    make([]*MetricsEvent, 0),
		sessionID: uuid.New().String(),
		clientID:  getOrGenerateClientID(),
	}
	exitCode := 123
	LogComplete(exitCode)

	if len(collector.events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(collector.events))
	}

	event := collector.events[0]
	if event.Type != "commands" {
		t.Errorf("Expected event type 'commands', got '%s'", event.Type)
	}
	if event.Name != "complete" {
		t.Errorf("Expected event name 'complete', got '%s'", event.Name)
	}
	if event.Metadata["exit_code"] != "123" {
		t.Errorf("Expected exit_code '123', got '%s'", event.Metadata["exit_code"])
	}
}

func TestGetOrGenerateClientID(t *testing.T) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("Failed to get user config dir: %v", err)
	}
	configDir = filepath.Join(configDir, configDirName)
	idFile := filepath.Join(configDir, idFileName)

	// Clean up any previous test runs
	os.RemoveAll(configDir)

	// First call should generate and save a new ID
	id1 := getOrGenerateClientID()
	if _, err := os.Stat(idFile); err != nil {
		t.Errorf("Expected id file to be created, but it wasn't: %v", err)
	}

	// Second call should return the same ID
	id2 := getOrGenerateClientID()
	if id1 != id2 {
		t.Errorf("Expected same ID, got '%s' and '%s'", id1, id2)
	}

	// Clean up after the test
	os.RemoveAll(configDir)
}

func TestGeneratePayload(t *testing.T) {
	collector = &MetricsCollector{
		mu:        sync.Mutex{},
		enabled:   true,
		events:    make([]*MetricsEvent, 0),
		sessionID: "test-session-id",
		clientID:  "test-client-id",
	}

	// Add a mock event
	collector.logEvent("test-type", "test-name", map[string]string{"test-key": "test-value"})

	payload, err := collector.generatePayload()
	if err != nil {
		t.Fatalf("generatePayload failed: %v", err)
	}

	// Check the generated payload (basic check, can be expanded)
	if len(payload) == 0 {
		t.Error("Payload is empty")
	}

	var logRequest map[string]interface{}
	err = json.Unmarshal(payload, &logRequest)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	logEvents, ok := logRequest["log_event"].([]interface{})
	if !ok || len(logEvents) != 1 {
		t.Fatalf("Expected 1 log event, got %d", len(logEvents))
	}

	logEvent := logEvents[0].(map[string]interface{})
	sourceExtensionJSON, ok := logEvent["source_extension_json"].(string)
	if !ok {
		t.Fatalf("Expected source_extension_json to be a string")
	}

	var concordEvent map[string]interface{}
	err = json.Unmarshal([]byte(sourceExtensionJSON), &concordEvent)
	if err != nil {
		t.Fatalf("Failed to unmarshal source_extension_json: %v", err)
	}

	if concordEvent["event_type"] != "test-type" {
		t.Errorf("Expected event type 'test-type', got '%s'", concordEvent["event_type"])
	}
	if concordEvent["event_name"] != "test-name" {
		t.Errorf("Expected event name 'test-name', got '%s'", concordEvent["event_name"])
	}
}

func TestMapToMetadataList(t *testing.T) {
	inputMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	metadataList := mapToMetadataList(inputMap)

	if len(metadataList) != 2 {
		t.Fatalf("Expected 2 metadata entries, got %d", len(metadataList))
	}

	// Check if the keys and values are correctly mapped
	foundKey1 := false
	foundKey2 := false
	for _, entry := range metadataList {
		if entry["key"] == "key1" && entry["value"] == "value1" {
			foundKey1 = true
		}
		if entry["key"] == "key2" && entry["value"] == "value2" {
			foundKey2 = true
		}
	}

	if !foundKey1 {
		t.Error("Missing or incorrect metadata for key1")
	}
	if !foundKey2 {
		t.Error("Missing or incorrect metadata for key2")
	}
}

func TestFlushDisabled(t *testing.T) {
	collector = &MetricsCollector{
		mu:      sync.Mutex{},
		enabled: false, // Disabled collector
		events: []*MetricsEvent{
			{Time: time.Now(), Type: "test", Name: "test", Metadata: map[string]string{"test": "test"}},
		},
		sessionID: "test-session-id",
		clientID:  "test-client-id",
	}
	// This should not panic or cause any errors, even with events present
	Flush()
	// Check if the events slice is still populated after Flush() is called
	if len(collector.events) != 1 {
		t.Errorf("Expected events slice to remain unchanged when telemetry is disabled, got %d events", len(collector.events))
	}
}

func TestGeneratePayloadEmptyEvents(t *testing.T) {
	collector = &MetricsCollector{
		mu:        sync.Mutex{},
		enabled:   true,
		events:    []*MetricsEvent{}, // Empty events slice
		sessionID: "test-session-id",
		clientID:  "test-client-id",
	}

	payload, err := collector.generatePayload()
	if err != nil {
		t.Fatalf("generatePayload failed: %v", err)
	}

	var logRequest map[string]interface{}
	err = json.Unmarshal(payload, &logRequest)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	logEvents, ok := logRequest["log_event"].([]interface{})
	if !ok || len(logEvents) != 0 {
		t.Fatalf("Expected 0 log events, got %d", len(logEvents))
	}
}
