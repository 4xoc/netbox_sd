// MIT License
//
// Copyright (c) 2024 WIIT AG
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
// Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
// WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
// OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package netbox

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCFInterface(t *testing.T) {
	assert.Implements(t, (*CustomFieldMap)(nil), CFMap{})
}

func TestCustomFieldJSON(t *testing.T) {
	var (
		cf       *CustomField
		testBool bool
		data     = []struct {
			src      string
			expected CFMap
		}{
			{"{}", CFMap{entries: make(map[string]*CustomField)}},
			{
				"{\"no_bgp\": true}",
				CFMap{
					entries: map[string]*CustomField{
						"no_bgp": &CustomField{CustomFieldBool, true},
					},
				},
			},
			{
				"{\"no_dhcp\":true,\"someInt\":123,\"some_text\":\"foobar\"}",
				CFMap{
					entries: map[string]*CustomField{
						"no_dhcp":   &CustomField{CustomFieldBool, true},
						"someInt":   &CustomField{CustomFieldNumber, float64(123)},
						"some_text": &CustomField{CustomFieldText, "foobar"},
					},
				},
			},
		}
		i      int
		err    error
		actual *CFMap
	)

	for i = range data {
		actual = new(CFMap)
		err = json.Unmarshal([]byte(data[i].src), actual)
		assert.NoError(t, err)
		assert.Equal(t, data[i].expected, *actual)
	}

	cf = data[i].expected.GetEntry("no_dhcp")
	assert.Equal(t, cf, data[i].expected.entries["no_dhcp"])

	_, err = cf.AsFloat()
	assert.ErrorIs(t, err, ErrCFCantConvertValue)

	_, err = cf.AsString()
	assert.ErrorIs(t, err, ErrCFCantConvertValue)

	testBool, err = cf.AsBool()
	assert.NoError(t, err)
	assert.Equal(t, testBool, data[i].expected.entries["no_dhcp"].Value.(bool))
}
