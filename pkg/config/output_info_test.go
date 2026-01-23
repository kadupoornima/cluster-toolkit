// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"testing"

	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v3"
)

func TestOutputInfo(t *testing.T) {
	TestingT(t)
}

// module outputs can be specified as a simple string for the output name or as
// a YAML mapping of name/description/sensitive (str,str,bool)
func (s *zeroSuite) TestUnmarshalOutputInfo(c *C) {
	var oinfo OutputInfo
	var y string

	y = "foo"
	c.Check(yaml.Unmarshal([]byte(y), &oinfo), IsNil)
	c.Check(oinfo, DeepEquals, OutputInfo{Name: "foo", Description: "", Sensitive: false})

	y = "{ name: foo }"
	c.Check(yaml.Unmarshal([]byte(y), &oinfo), IsNil)
	c.Check(oinfo, DeepEquals, OutputInfo{Name: "foo", Description: "", Sensitive: false})

	y = "{ name: foo, description: bar }"
	c.Check(yaml.Unmarshal([]byte(y), &oinfo), IsNil)
	c.Check(oinfo, DeepEquals, OutputInfo{Name: "foo", Description: "bar", Sensitive: false})

	y = "{ name: foo, description: bar, sensitive: true }"
	c.Check(yaml.Unmarshal([]byte(y), &oinfo), IsNil)
	c.Check(oinfo, DeepEquals, OutputInfo{Name: "foo", Description: "bar", Sensitive: true})

	// extra key should generate error
	y = "{ name: foo, description: bar, sensitive: true, extrakey: extraval }"
	c.Check(yaml.Unmarshal([]byte(y), &oinfo), NotNil)

	// missing required key name should generate error
	y = "{ description: bar, sensitive: true }"
	c.Check(yaml.Unmarshal([]byte(y), &oinfo), NotNil)

	// should not ummarshal a sequence
	y = "[ foo ]"
	c.Check(yaml.Unmarshal([]byte(y), &oinfo), NotNil)

	// should not ummarshal an object with non-boolean sensitive type
	y = "{ name: foo, description: bar, sensitive: contingent }"
	c.Check(yaml.Unmarshal([]byte(y), &oinfo), NotNil)
}
