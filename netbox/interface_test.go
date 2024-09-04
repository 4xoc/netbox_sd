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
	iface1 = &Interface{
		ID:        1,
		IDString:  "1",
		Name:      "ipmi",
		Enabled:   true,
		isVirtual: false,
		CustomFields: CFMap{
			entries: map[string]*CustomField{
				"custom_field_E": &CustomField{
					Datatype: "text",
					Value:    "foobar",
				},
				"custom_field_F": &CustomField{
					Datatype: "integer",
					Value:    float64(123),
				},
				"custom_field_G": &CustomField{
					Datatype: "boolean",
					Value:    false,
				},
			},
		},
		Tags: []Name{
			{
				Name: "ipmi_exporter",
			},
		},
		Device: devA,
	}
	iface2 = &Interface{
		ID:        2,
		IDString:  "2",
		Name:      "ipmi",
		Enabled:   true,
		isVirtual: false,
		CustomFields: CFMap{
			entries: map[string]*CustomField{},
		},
		Tags: []Name{
			{
				Name: "ipmi_exporter",
			},
		},
		Device: devB,
	}

	vIface1 = &Interface{
		ID:        1,
		IDString:  "1",
		Name:      "eth0",
		Enabled:   true,
		isVirtual: true,
		CustomFields: CFMap{
			entries: map[string]*CustomField{
				"custom_field_K": &CustomField{
					Datatype: "text",
					Value:    "ytrewq",
				},
				"custom_field_L": &CustomField{
					Datatype: "integer",
					Value:    float64(999),
				},
				"custom_field_M": &CustomField{
					Datatype: "boolean",
					Value:    true,
				},
			},
		},
		Tags: []Name{
			{
				Name: "node_exporter",
			},
		},
		Device: vmA,
	}
	vIface2 = &Interface{
		ID:        2,
		IDString:  "2",
		Name:      "eth0",
		Enabled:   true,
		isVirtual: true,
		CustomFields: CFMap{
			entries: map[string]*CustomField{},
		},
		Tags: []Name{
			{
				Name: "node_exporter",
			},
		},
		Device: vmB,
	}
)

func TestGetInterface(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	// interface exists
	iface, err := client.GetInterface(1)
	assert.NoError(t, err)
	require.NotEmpty(t, iface)

	assert.Equal(t, iface1, iface)

	// interface is missing
	iface, err = client.GetInterface(99999)
	assert.NoError(t, err)
	assert.Empty(t, iface)
}

func TestGetInterfacesByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	client := newTestClient(t)

	// interface exists
	iface, err := client.GetInterfacesByTag("ipmi_exporter")
	assert.NoError(t, err)
	require.NotEmpty(t, iface)
	require.Equal(t, 2, len(iface))

	assert.Equal(t, []*Interface{iface1, iface2}, iface)

	// interface is missing
	iface, err = client.GetInterfacesByTag("this-does_not-exist")
	assert.NoError(t, err)
	require.Empty(t, iface)
}

func TestGetVirtualInterface(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	// interface exists
	iface, err := client.GetVirtualInterface(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, iface)
	assert.Equal(t, vIface1, iface)

	// interface is missing
	iface, err = client.GetVirtualInterface(99999)
	assert.NoError(t, err)
	assert.Empty(t, iface)
}

func TestGetVirtualInterfacesByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	// interface exists
	iface, err := client.GetVirtualInterfacesByTag("node_exporter")
	assert.NoError(t, err)
	require.NotEmpty(t, iface)
	require.Equal(t, 2, len(iface))

	assert.Equal(t, []*Interface{vIface1, vIface2}, iface)

	// interface is missing
	iface, err = client.GetVirtualInterfacesByTag("this-does_not-exist")
	assert.NoError(t, err)
	require.Empty(t, iface)
}
