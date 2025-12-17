// Copyright 2024 "Google LLC"
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
	"errors"
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

// SetCLIVariables parses "key=value" strings and adds them to the DeploymentSettings variables.
func SetCLIVariables(ds *DeploymentSettings, s []string) error {
	for _, cliVar := range s {
		arr := strings.SplitN(cliVar, "=", 2)

		if len(arr) != 2 {
			return fmt.Errorf("invalid format: '%s' should follow the 'name=value' format", cliVar)
		}
		// Convert the variable's string literal to its equivalent default type.
		key := arr[0]
		var v YamlValue
		if err := yaml.Unmarshal([]byte(arr[1]), &v); err != nil {
			return fmt.Errorf("invalid input: unable to convert '%s' value '%s' to known type", key, arr[1])
		}
		ds.Vars = ds.Vars.With(key, v.Unwrap())
	}
	return nil
}

// SetBackendConfig parses "key=value" strings and configures the TerraformBackend.
func SetBackendConfig(ds *DeploymentSettings, s []string) error {
	if len(s) == 0 {
		return nil // no op
	}
	be := TerraformBackend{Type: "gcs"}
	for _, config := range s {
		arr := strings.SplitN(config, "=", 2)

		if len(arr) != 2 {
			return fmt.Errorf("invalid format: '%s' should follow the 'name=value' format", config)
		}

		key, value := arr[0], arr[1]
		switch key {
		case "type":
			be.Type = value
		default:
			be.Configuration = be.Configuration.With(key, cty.StringVal(value))
		}
	}
	ds.TerraformBackendDefaults = be
	return nil
}

// SetValidationLevel sets the validation level for the blueprint.
func SetValidationLevel(bp *Blueprint, s string) error {
	switch s {
	case "ERROR":
		bp.ValidationLevel = ValidationError
	case "WARNING":
		bp.ValidationLevel = ValidationWarning
	case "IGNORE":
		bp.ValidationLevel = ValidationIgnore
	default:
		return errors.New("invalid validation level (\"ERROR\", \"WARNING\", \"IGNORE\")")
	}
	return nil
}

// SkipValidators marks specified validators to be skipped.
func SkipValidators(bp *Blueprint, validatorsToSkip []string) {
	for _, v := range validatorsToSkip {
		bp.SkipValidator(v)
	}
}
