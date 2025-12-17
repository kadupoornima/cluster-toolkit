// Copyright 2023 Google LLC
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
	"fmt"
	"gopkg.in/yaml.v3"
)

// ModuleOutput defines a module output in the blueprint
type ModuleOutput struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Sensitive   bool   `yaml:"sensitive,omitempty"`
}

// UnmarshalYAML supports parsing YAML ModuleOutput fields as a simple list of
// strings or as a list of maps directly into ModuleOutput struct
func (mo *ModuleOutput) UnmarshalYAML(value *yaml.Node) error {
	var name string
	const yamlErrorMsg string = "block beginning at line %d: %s"

	err := value.Decode(&name)
	if err == nil {
		mo.Name = name
		return nil
	}

	var fields map[string]interface{}
	err = value.Decode(&fields)
	if err != nil {
		return fmt.Errorf(yamlErrorMsg, value.Line, "outputs must each be a string or a map{name: string, description: string, sensitive: bool}; "+err.Error())
	}

	err = enforceMapKeys(fields, map[string]bool{
		"name": true, "description": false, "sensitive": false},
	)
	if err != nil {
		return fmt.Errorf(yamlErrorMsg, value.Line, err)
	}

	type rawModuleOutput ModuleOutput
	if err := value.Decode((*rawModuleOutput)(mo)); err != nil {
		return fmt.Errorf("line %d: %s", value.Line, err)
	}
	return nil
}

// enforceMapKeys ensures the presence of required keys and absence of unallowed
// keys with a useful error message; input is a map of all allowed keys to a
// boolean that is true when key is required and false when optional
func enforceMapKeys(input map[string]interface{}, allowedKeys map[string]bool) error {
	for key := range input {
		if _, ok := allowedKeys[key]; !ok {
			return fmt.Errorf("provided invalid key: %#v", key)
		}
		allowedKeys[key] = false
	}
	for key, req := range allowedKeys {
		if req {
			return fmt.Errorf("missing required key: %#v", key)
		}
	}
	return nil
}
