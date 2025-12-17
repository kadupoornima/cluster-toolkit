package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// OutputInfo stores information about module output values
type OutputInfo struct {
	Name        string
	Description string `yaml:",omitempty"`
	Sensitive   bool   `yaml:",omitempty"`
	// DependsOn   []string `yaml:"depends_on,omitempty"`
}

// UnmarshalYAML supports parsing YAML OutputInfo fields as a simple list of
// strings or as a list of maps directly into OutputInfo struct
func (mo *OutputInfo) UnmarshalYAML(value *yaml.Node) error {
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

	type rawOutputInfo OutputInfo
	if err := value.Decode((*rawOutputInfo)(mo)); err != nil {
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
