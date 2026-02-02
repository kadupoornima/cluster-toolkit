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
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestTriggerBackgroundUpload(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-telemetry-upload-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Call triggerBackgroundUpload with the temporary file's path
	triggerBackgroundUpload(tmpFile.Name())

	// Give the background process some time to start (very rough estimation)
	// In a real test, you might want to use a more reliable synchronization mechanism.
	// time.Sleep(100 * time.Millisecond)

	// Check if the background process was started correctly by looking for a process
	// with the expected command-line arguments.  Note: This is inherently racy and
	// fragile.  A better approach would involve more direct process management (if possible).
	found := false
	processes, err := getProcesses()
	if err != nil {
		t.Fatalf("Failed to get processes: %v", err)
	}
	executablePath, err := os.Executable()
	if err != nil {
		t.Fatalf("Failed to get executable path: %v", err)
	}

	for _, process := range processes {
		if strings.Contains(process, executablePath) && strings.Contains(process, "internal-emit-telemetry") && strings.Contains(process, tmpFile.Name()) {
			found = true
			break
		}
	}

	// Report an error if we didn't find the process
	if !found {
		t.Errorf("Background process not found.  Expected process with args '%s internal-emit-telemetry %s'", executablePath, tmpFile.Name())
	}
}

// getProcesses is a helper function to get a list of running processes.
// Note: This implementation is OS-specific and may need to be adjusted based on the target OS.
func getProcesses() ([]string, error) {
	cmd := exec.Command("ps", "aux") // Linux/macOS
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(output), "\n"), nil
}
