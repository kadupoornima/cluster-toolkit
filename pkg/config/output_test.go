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
	"gopkg.in/yaml.v3"

	. "gopkg.in/check.v1"
)

func (s *zeroSuite) TestUnmarshalOutput(c *C) {
	var out Output
	var y string

	y = "foo"
	c.Check(yaml.Unmarshal([]byte(y), &out), IsNil)
	c.Check(out, DeepEquals, Output{Name: "foo", Description: "", Sensitive: false})

	y = "{ name: foo }"
	c.Check(yaml.Unmarshal([]byte(y), &out), IsNil)
	c.Check(out, DeepEquals, Output{Name: "foo", Description: "", Sensitive: false})

	y = "{ name: foo, description: bar }"
	c.Check(yaml.Unmarshal([]byte(y), &out), IsNil)
	c.Check(out, DeepEquals, Output{Name: "foo", Description: "bar", Sensitive: false})

	y = "{ name: foo, description: bar, sensitive: true }"
	c.Check(yaml.Unmarshal([]byte(y), &out), IsNil)
	c.Check(out, DeepEquals, Output{Name: "foo", Description: "bar", Sensitive: true})

	// Fail
	y = "foo: bar"
	c.Check(yaml.Unmarshal([]byte(y), &out), NotNil)

	y = "{ name: foo, extra: key }"
	c.Check(yaml.Unmarshal([]byte(y), &out), NotNil)
}
