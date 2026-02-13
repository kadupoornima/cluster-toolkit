package validators

import (
	"google.golang.org/api/googleapi"
	"gopkg.in/check.v1"
)

type CloudSuite struct{}

var _ = check.Suite(&CloudSuite{})

func (s *CloudSuite) TestGetErrorReason(c *check.C) {
	// Case 1: Valid reason and metadata
	err1 := googleapi.Error{
		Details: []interface{}{
			map[string]interface{}{
				"reason": "SOME_REASON",
				"metadata": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}
	reason, metadata := getErrorReason(err1)
	c.Assert(reason, check.Equals, "SOME_REASON")
	c.Assert(metadata, check.NotNil)
	c.Assert(metadata["key"], check.Equals, "value")

	// Case 2: Valid reason, missing metadata (Should NOT panic)
	err2 := googleapi.Error{
		Details: []interface{}{
			map[string]interface{}{
				"reason": "ANOTHER_REASON",
				// "metadata" is missing
			},
		},
	}
	reason, metadata = getErrorReason(err2)
	c.Assert(reason, check.Equals, "ANOTHER_REASON")
	c.Assert(metadata, check.IsNil)

	// Case 3: Valid reason, nil metadata (Should NOT panic)
	err3 := googleapi.Error{
		Details: []interface{}{
			map[string]interface{}{
				"reason":   "YET_ANOTHER_REASON",
				"metadata": nil,
			},
		},
	}
	reason, metadata = getErrorReason(err3)
	c.Assert(reason, check.Equals, "YET_ANOTHER_REASON")
	c.Assert(metadata, check.IsNil)

	// Case 4: Valid reason, invalid metadata type (Should NOT panic)
	err4 := googleapi.Error{
		Details: []interface{}{
			map[string]interface{}{
				"reason":   "BAD_METADATA",
				"metadata": "not_a_map",
			},
		},
	}
	reason, metadata = getErrorReason(err4)
	c.Assert(reason, check.Equals, "BAD_METADATA")
	c.Assert(metadata, check.IsNil)

	// Case 5: No reason found
	err5 := googleapi.Error{
		Details: []interface{}{
			map[string]interface{}{
				"other": "field",
			},
		},
	}
	reason, metadata = getErrorReason(err5)
	c.Assert(reason, check.Equals, "")
	c.Assert(metadata, check.IsNil)
}
