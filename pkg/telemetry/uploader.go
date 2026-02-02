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
	"log"
	"os"
	"os/exec"
)

func triggerBackgroundUpload(filePath string) {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		// If we can't find ourselves, fallback to Args[0] or fail silently
		exe = os.Args[0]
	}

	cmd := exec.Command(exe, "internal-emit-telemetry", filePath)

	// Detach process: start it and release resources immediately
	if err := cmd.Start(); err != nil {
		return
	}
	// Release allows the child to outlive the parent
	_ = cmd.Process.Release()
}
