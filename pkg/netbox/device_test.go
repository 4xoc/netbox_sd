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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	devA = &Device{
		ID:         1,
		IDString:   "1",
		Name:       "device-A",
		PrimaryIP4: ip2,
		PrimaryIP6: ip1,
		CustomFields: CFMap{
			entries: map[string]*CustomField{
				"custom_field_A": &CustomField{
					Datatype: "text",
					Value:    "bla",
				},
				"custom_field_B": &CustomField{
					Datatype: "integer",
					Value:    float64(1),
				},
				"custom_field_C": &CustomField{
					Datatype: "boolean",
					Value:    true,
				},
			},
		},
		Rack: Name{
			Name: "site-A-rack-A",
		},
		Site: Name{
			Name: "site-A",
		},
		Role: Name{
			Name: "role-A",
		},
		Tenant: Name{
			Name: "tenant-A",
		},
		Platform: Name{
			Name: "platform-A",
		},
		SerialNumber: "abcd",
		AssetTag:     "a1234",
		Status:       "ACTIVE",
		Tags: []Name{
			{
				Name: "node_exporter",
			},
		},
		isVirtual: false,
	}
	devB = &Device{
		ID:         2,
		IDString:   "2",
		Name:       "device-B",
		PrimaryIP4: nil,
		PrimaryIP6: nil,
		CustomFields: CFMap{
			entries: map[string]*CustomField{},
		},
		Rack: Name{
			Name: "site-B-rack-A",
		},
		Site: Name{
			Name: "site-B",
		},
		Role: Name{
			Name: "role-B",
		},
		Tenant: Name{
			Name: "tenant-B",
		},
		Platform: Name{
			Name: "platform-B",
		},
		SerialNumber: "abcde",
		AssetTag:     "a12345",
		Status:       "ACTIVE",
		Tags: []Name{
			{
				Name: "node_exporter",
			},
		},
		isVirtual: false,
	}
)

func TestGetDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	// device exists
	dev, err := client.GetDevice(1)
	assert.NoError(t, err)
	require.NotEmpty(t, dev)
	assert.Equal(t, devA, dev)

	// device is missing
	dev, err = client.GetDevice(99999)
	assert.NoError(t, err)
	assert.Empty(t, dev)
}

func TestGetDevices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	devs, err := client.GetDevices()
	assert.NoError(t, err)
	require.Len(t, devs, 2)

	// validating contents
	assert.Equal(t, []*Device{devA, devB}, devs)
}

func TestGetDevicesByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	devs, err := client.GetDevicesByTag("node_exporter")
	assert.NoError(t, err)
	require.Len(t, devs, 2)

	// validating contents
	assert.Equal(t, []*Device{devA, devB}, devs)

	// tag doesn't exist
	devs, err = client.GetDevicesByTag("doesn_t-exist")
	assert.NoError(t, err)
	assert.Empty(t, devs)
}
