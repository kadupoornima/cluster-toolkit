// Copyright 2026 Google LLC
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
	"fmt"
	"hpc-toolkit/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(telemetryCmd)
}

var telemetryCmd = &cobra.Command{
	Use:   "telemetry [on|off]",
	Short: "Enable or disable telemetry",
	Args:  cobra.ExactArgs(1), // Ensure exactly one argument is provided
	RunE: func(cmd *cobra.Command, args []string) error {
		val := args[0]
		var enabled bool

		// 1. Logic to parse "on/off" or "true/false"
		switch val {
		case "on", "true", "yes":
			enabled = true
		case "off", "false", "no":
			enabled = false
		default:
			return fmt.Errorf("invalid argument %q: use 'on' or 'off'", val)
		}

		// 2. Update Viper (Memory)
		viper.Set(config.TELEMETRY_KEY, enabled)

		// 3. Persist to Firestore (Remote)
		err := config.SaveToFirestore()
		if err != nil {
			return fmt.Errorf("could not save setting to cloud: %w", err)
		}

		fmt.Printf("Telemetry has been turned %s.\n", val)
		return nil
	},
}
