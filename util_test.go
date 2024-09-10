// MIT License
//
// Copyright (c) 2024 WIIT AG
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit
// persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
// Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
// WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
// OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"testing"

	"github.com/4xoc/netbox_sd/internal/config"
	"github.com/4xoc/netbox_sd/internal/util"
	"github.com/4xoc/netbox_sd/pkg/netbox"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// Mocks
//

type cfMap struct {
	netbox.CustomFieldMap
	entries map[string]*netbox.CustomField
}

func (cfm cfMap) GetEntry(name string) *netbox.CustomField {
	var (
		val *netbox.CustomField
		ok  bool
	)

	if val, ok = cfm.entries[name]; !ok {
		return nil
	}

	return val
}

func (cfm cfMap) GetAllEntries(callback func(string, *netbox.CustomField)) {
	var key string

	for key = range cfm.entries {
		callback(key, cfm.entries[key])
	}
}

//
// Tests
//

func TestSelectAddr(t *testing.T) {
	var (
		data = []struct {
			input    []*netbox.IP
			group    *config.Group
			expected []*netbox.IP
		}{
			{
				input: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("inet6"),
						AllAddresses: util.NewPtr[bool](true),
					},
				},
				expected: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
			},
			{
				// inetFamily not matching
				input: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("inet"),
						AllAddresses: util.NewPtr[bool](true),
					},
				},
				expected: []*netbox.IP{},
			},
			{
				// only all inet6 addresses and nil addr
				input: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::123",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPActive,
					},
					nil,
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("inet6"),
						AllAddresses: util.NewPtr[bool](true),
					},
				},
				expected: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::123",
						Status:  netbox.StatusIPActive,
					},
				},
			},
			{
				// only all inet6 addresses with duplicates
				input: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::123",
						Status:  netbox.StatusIPSLAAC,
					},
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPActive,
					},
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("inet6"),
						AllAddresses: util.NewPtr[bool](true),
					},
				},
				expected: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::123",
						Status:  netbox.StatusIPSLAAC,
					},
				},
			},
			{
				// only one inet address
				input: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPDHCP,
					},
					&netbox.IP{
						Address: "10.0.0.1",
						Status:  netbox.StatusIPActive,
					},
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("inet"),
						AllAddresses: util.NewPtr[bool](false),
					},
				},
				expected: []*netbox.IP{
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPDHCP,
					},
				},
			},
			{
				// only one inet6 address
				input: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::123",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::12",
						Status:  netbox.StatusIPActive,
					},
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("inet6"),
						AllAddresses: util.NewPtr[bool](false),
					},
				},
				expected: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
			},
			{
				// all any inet
				input: []*netbox.IP{
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "10.0.0.1",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("any"),
						AllAddresses: util.NewPtr[bool](true),
					},
				},
				expected: []*netbox.IP{
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "10.0.0.1",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
			},
			{
				// nil ptr shouldn't panic
				input: []*netbox.IP{
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "10.0.0.1",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("any"),
						AllAddresses: util.NewPtr[bool](true),
					},
				},
				expected: []*netbox.IP{
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "10.0.0.1",
						Status:  netbox.StatusIPActive,
					},
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
			},
			{
				// non active status must be honored
				input: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPReserved,
					},
					&netbox.IP{
						Address: "10.0.0.0",
						Status:  netbox.StatusIPDeprecated,
					},
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
				group: &config.Group{
					Flags: config.Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string]("any"),
						AllAddresses: util.NewPtr[bool](true),
					},
				},
				expected: []*netbox.IP{
					&netbox.IP{
						Address: "2001:db8::1234",
						Status:  netbox.StatusIPActive,
					},
				},
			},
		}
		result []*netbox.IP
		i      int
	)

	for i = range data {
		result = selectAddr(data[i].input, data[i].group)
		assert.Equal(t, data[i].expected, result)
	}
}

func TestGenerateCustomFieldLabels(t *testing.T) {
	var (
		input netbox.CustomFieldMap = cfMap{
			entries: map[string]*netbox.CustomField{
				"foo": &netbox.CustomField{
					Datatype: netbox.CustomFieldText,
					Value:    "bar",
				},
				"foo2": &netbox.CustomField{
					Datatype: netbox.CustomFieldNumber,
					Value:    float64(123),
				},
				"foo3": &netbox.CustomField{
					Datatype: netbox.CustomFieldBool,
					Value:    true,
				},
			},
		}
		expected model.LabelSet = model.LabelSet{
			"netbox_foo":  "bar",
			"netbox_foo2": "123",
			"netbox_foo3": "true",
		}
		result model.LabelSet
		err    error
	)

	result, err = generateCustomFieldLabels(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
