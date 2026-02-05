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

package config

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// UserConfigFilename is the name of the user configuration file
const UserConfigFilename = "config.yaml"

// UserConfigDirName is the directory name for cluster-toolkit config
const UserConfigDirName = "cluster-toolkit"

// UserConfig stores persistent user settings
type UserConfig struct {
	ClientID string `yaml:"client_id"`
}

// GetUserConfigPath returns the full path to the user configuration file
func GetUserConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, UserConfigDirName, UserConfigFilename), nil
}

// LoadUserConfig reads the configuration from the file
func LoadUserConfig() (UserConfig, error) {
	path, err := GetUserConfigPath()
	if err != nil {
		return UserConfig{}, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return UserConfig{}, nil // Return empty config if file doesn't exist
	}
	if err != nil {
		return UserConfig{}, err
	}

	var cfg UserConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return UserConfig{}, err
	}
	return cfg, nil
}

// SaveUserConfig writes the configuration to the file
func SaveUserConfig(cfg UserConfig) error {
	path, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetPersistentClientID returns the stored Client ID or generates/saves a new one
func GetPersistentClientID() (string, error) {
	cfg, err := LoadUserConfig()
	if err != nil {
		return "", err
	}

	// If ClientID is missing, generate a new one and save it
	if cfg.ClientID == "" {
		cfg.ClientID = uuid.NewString()
		if err := SaveUserConfig(cfg); err != nil {
			return "", err
		}
	}

	return cfg.ClientID, nil
}
