/**
 * Copyright 2026 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package modulereader extracts necessary information from modules
package modulereader

import (
	"fmt"
	"hpc-toolkit/pkg/logging"
	"hpc-toolkit/pkg/sourcereader"
	"os"
	"path"

	"github.com/hashicorp/go-getter"
	"github.com/zclconf/go-cty/cty"
)

// VarInfo stores information about a module input variables
type VarInfo struct {
	Name        string
	Type        cty.Type
	Description string
	Default     interface{}
	Required    bool
}

// OutputInfo stores information about module output values
type OutputInfo struct {
	Name        string
	Description string `yaml:",omitempty"`
	Sensitive   bool   `yaml:",omitempty"`
	// DependsOn   []string `yaml:"depends_on,omitempty"`
}

// ModuleInfo stores information about a module
type ModuleInfo struct {
	Inputs   []VarInfo
	Outputs  []OutputInfo
	Metadata Metadata
}

// GetOutputsAsMap returns the outputs list as a map for quicker access
func (i ModuleInfo) GetOutputsAsMap() map[string]OutputInfo {
	outputsMap := make(map[string]OutputInfo)
	for _, output := range i.Outputs {
		outputsMap[output.Name] = output
	}
	return outputsMap
}

type sourceAndKind struct {
	source string
	kind   string
}

var modInfoCache = map[sourceAndKind]ModuleInfo{}
var modDownloadCache = map[string]string{} // Cache for downloaded module data

// GetModuleInfo gathers information about a module at a given source using the
// tfconfig package. It will add details about required APIs to be
// enabled for that module.
// There is a cache to avoid re-reading the module info for the same source and kind.
func GetModuleInfo(source string, kind string) (ModuleInfo, error) {
	key := sourceAndKind{source, kind}
	if mi, ok := modInfoCache[key]; ok {
		return mi, nil
	}

	var modPath string
	switch {
	case sourcereader.IsEmbeddedPath(source) || sourcereader.IsLocalPath(source):
		modPath = source
		if sourcereader.IsLocalPath(source) && sourcereader.LocalModuleIsEmbedded(source) {
			return ModuleInfo{}, fmt.Errorf("using embedded modules with local paths is no longer supported; use embedded path and rebuild gcluster binary")
		}
	default:
		pkgAddr, subDir := getter.SourceDirSubdir(source)
		if cachedModPath, ok := modDownloadCache[pkgAddr]; ok {
			modPath = path.Join(cachedModPath, subDir)
		} else {
			tmpDir, err := os.MkdirTemp("", "module-*")
			if err != nil {
				return ModuleInfo{}, err
			}

			pkgPath := path.Join(tmpDir, "module")
			modPath = path.Join(pkgPath, subDir)
			sourceReader := sourcereader.Factory(pkgAddr)
			if err = sourceReader.GetModule(pkgAddr, pkgPath); err != nil {
				if subDir != "" && kind == "packer" {
					err = fmt.Errorf("module source %s included \"//\" package syntax; "+
						"the \"//\" should typically be placed at the root of the repository:\n%w", source, err)
				}
				return ModuleInfo{}, err
			}
			modDownloadCache[pkgAddr] = pkgPath
		}
	}

	reader := Factory(kind)
	mi, err := reader.GetInfo(modPath)
	if err != nil {
		return ModuleInfo{}, err
	}
	mi.Metadata = GetMetadataSafe(modPath)
	modInfoCache[key] = mi
	return mi, nil
}

// SetModuleInfo sets the ModuleInfo for a given source and kind
// NOTE: This is only used for testing
func SetModuleInfo(source string, kind string, info ModuleInfo) {
	modInfoCache[sourceAndKind{source, kind}] = info
}

// ModReader is a module reader interface
type ModReader interface {
	GetInfo(path string) (ModuleInfo, error)
}

var kinds = map[string]ModReader{
	"terraform": NewTFReader(),
	"packer":    NewPackerReader(),
}

// Factory returns a ModReader of type 'kind'
func Factory(kind string) ModReader {
	r, ok := kinds[kind]
	if !ok {
		logging.Fatal("Invalid request to create a reader of kind %s", kind)
	}
	return r
}
