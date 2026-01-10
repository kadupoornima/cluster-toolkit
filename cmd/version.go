/*
Copyright 2022 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"hpc-toolkit/pkg/logging"
	"hpc-toolkit/pkg/shell"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gcluster",
	Long:  `All software has versions. This is gcluster's.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use the logic from root.go to populate annotations
		// But here we print it directly.

		// Note: root.go's Execute() populates annotations based on GitCommitInfo
		// If we run `go run main.go`, GitCommitInfo might be empty unless ldflags are used.
		// However, rootCmd.Version is hardcoded in root.go as well.

		// Let's print the hardcoded version first
		logging.Info("gcluster version: %s", rootCmd.Version)

		// Print git info if available
		if val, ok := annotation["branch"]; ok {
			logging.Info("Built from branch: %s", val)
		}
		if val, ok := annotation["commitInfo"]; ok {
			logging.Info("Commit info: %s", val)
		}

		// Check Terraform version
		tfVersion, err := shell.TfVersion()
		if err == nil && tfVersion != "" {
			logging.Info("Terraform version: %s", tfVersion)
		} else {
			logging.Info("Terraform version: not found or error: %v", err)
		}

		// Check Packer version
		packerVersion, err := shell.PackerVersion()
		if err == nil && packerVersion != "" {
			logging.Info("Packer version: %s", packerVersion)
		} else {
			// Packer is not strictly required for all operations, so just log it's not found
			logging.Info("Packer version: not found")
		}
	},
}
