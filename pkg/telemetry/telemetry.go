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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

const (
	ClearcutURL   = "https://play.googleapis.com/log"
	LogSourceName = "CONCORD"
	ClientType    = "GCLUSTER"
	configDirName = "cluster-toolkit"
	idFileName    = "telemetry_id"
)

var collector *MetricsCollector

func init() {
	collector = &MetricsCollector{
		events:    make([]*MetricsEvent, 0),
		sessionID: uuid.New().String(),
		clientID:  getOrGenerateClientID(),
	}
}

// --- Public Interface ---

func Init(enabled bool) {
	collector.mu.Lock()
	defer collector.mu.Unlock()
	collector.enabled = enabled
}

func LogStart(command string) {
	collector.logEvent("commands", "start", map[string]string{
		"command": command,
	})
}

// LogComplete captures the exit code and logs the completion event
func LogComplete(exitCode int) {
	collector.logEvent("commands", "complete", map[string]string{
		"exit_code": fmt.Sprintf("%d", exitCode),
	})
}

func Flush() {
	collector.mu.Lock()
	if !collector.enabled || len(collector.events) == 0 {
		collector.mu.Unlock()
		return
	}
	payload, err := collector.generatePayload()
	collector.mu.Unlock() // Unlock before expensive IO operations

	if err != nil {
		return
	}

	// 1. Write payload to a temp file
	f, err := os.CreateTemp("", "gcluster-telemetry-*.json")
	if err != nil {
		return
	}
	defer f.Close()

	// Prepare the full upload configuration for the background worker
	uploadConfig := map[string]interface{}{
		"url":     ClearcutURL,
		"method":  "POST",
		"headers": map[string]string{"User-Agent": "gcluster-telemetry/1.0"},
		"data":    string(payload),
		"params":  map[string]string{"format": "json_proto"},
	}

	if err := json.NewEncoder(f).Encode(uploadConfig); err != nil {
		return
	}

	// 2. Spawn the detached background process
	// We call the current binary again with the hidden subcommand
	selfExe, err := os.Executable()
	if err != nil {
		return
	}

	cmd := exec.Command(selfExe, "internal-telemetry", f.Name())

	// CRITICAL: Ensure the child process is fully detached and doesn't hold open IO pipes
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	configureDetachedProcess(cmd)

	_ = cmd.Start()
}

// --- Internal Logic ---

type MetricsCollector struct {
	mu        sync.Mutex
	enabled   bool
	events    []*MetricsEvent
	sessionID string
	clientID  string
}

type MetricsEvent struct {
	Time     time.Time
	Type     string
	Name     string
	Metadata map[string]string
}

func (c *MetricsCollector) logEvent(eventType, name string, metadata map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return
	}

	c.events = append(c.events, &MetricsEvent{
		Time:     time.Now(),
		Type:     eventType,
		Name:     name,
		Metadata: metadata,
	})
}

func (c *MetricsCollector) generatePayload() ([]byte, error) {
	baseMetadata := map[string]string{
		"session_id": c.sessionID,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
	}

	serializedEvents := make([]map[string]interface{}, 0)
	var firstTime time.Time
	if len(c.events) > 0 {
		firstTime = c.events[0].Time
	}

	for _, e := range c.events {
		finalMeta := make(map[string]string)
		for k, v := range baseMetadata {
			finalMeta[k] = v
		}
		for k, v := range e.Metadata {
			finalMeta[k] = v
		}

		// Calculate latency relative to the start event
		if !firstTime.IsZero() {
			finalMeta["latency_seconds"] = fmt.Sprintf("%d", int(e.Time.Sub(firstTime).Seconds()))
		}

		concordEvent := map[string]interface{}{
			"console_type":      ClientType,
			"client_install_id": c.clientID,
			"event_type":        e.Type,
			"event_name":        e.Name,
			"event_metadata":    mapToMetadataList(finalMeta),
		}

		concordJson, _ := json.Marshal(concordEvent)

		serializedEvents = append(serializedEvents, map[string]interface{}{
			"event_time_ms":         e.Time.UnixMilli(),
			"source_extension_json": string(concordJson),
		})
	}

	logRequest := map[string]interface{}{
		"client_info":     map[string]string{"client_type": ClientType},
		"log_source_name": LogSourceName,
		"request_time_ms": time.Now().UnixMilli(),
		"log_event":       serializedEvents,
	}

	return json.Marshal(logRequest)
}

func getOrGenerateClientID() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return uuid.New().String() // Fallback if no home dir
	}

	fullPath := filepath.Join(configDir, configDirName, idFileName)

	// Try reading existing ID
	if data, err := os.ReadFile(fullPath); err == nil {
		return string(data)
	}

	// Generate and save new ID
	id := uuid.New().String()
	_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
	_ = os.WriteFile(fullPath, []byte(id), 0644)
	return id
}

func mapToMetadataList(m map[string]string) []map[string]string {
	var list []map[string]string
	for k, v := range m {
		list = append(list, map[string]string{"key": k, "value": v})
	}
	return list
}

func configureDetachedProcess(cmd *exec.Cmd) {
	if runtime.GOOS == "windows" {
		// cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `cmd /C "` + cmd.Path + `"`, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	}
}
