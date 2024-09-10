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
	vmA = &Device{
		ID:         1,
		IDString:   "1",
		Name:       "vm-A",
		PrimaryIP4: ip3,
		PrimaryIP6: ip4,
		CustomFields: CFMap{
			entries: map[string]*CustomField{
				"custom_field_H": &CustomField{
					Datatype: "text",
					Value:    "qwerty",
				},
				"custom_field_I": &CustomField{
					Datatype: "integer",
					Value:    float64(666),
				},
				"custom_field_J": &CustomField{
					Datatype: "boolean",
					Value:    false,
				},
			},
		},
		Site: Name{
			Name: "site-A",
		},
		Role: Name{
			Name: "role-A",
		},
		Tenant: Name{
			Name: "tenant-C",
		},
		Platform: Name{
			Name: "platform-A",
		},
		Status: "ACTIVE",
		Tags: []Name{
			{
				Name: "node_exporter",
			},
		},
		isVirtual: true,
	}
	vmB = &Device{
		ID:         2,
		IDString:   "2",
		Name:       "vm-B",
		PrimaryIP4: ip5,
		PrimaryIP6: ip6,
		CustomFields: CFMap{
			entries: map[string]*CustomField{
				"custom_field_H": &CustomField{
					Datatype: "text",
					Value:    "foobar",
				},
				"custom_field_I": &CustomField{
					Datatype: "integer",
					Value:    float64(9876),
				},
				"custom_field_J": &CustomField{
					Datatype: "boolean",
					Value:    true,
				},
			},
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
		Status: "ACTIVE",
		Tags: []Name{
			{
				Name: "node_exporter",
			},
		},
		isVirtual: true,
	}
	vmC = &Device{
		ID:         3,
		IDString:   "3",
		Name:       "vm-C",
		PrimaryIP4: nil,
		PrimaryIP6: nil,
		CustomFields: CFMap{
			entries: map[string]*CustomField{},
		},
		Site: Name{
			Name: "site-C",
		},
		Role: Name{
			Name: "role-C",
		},
		Status:    "ACTIVE",
		Tags:      []Name{},
		isVirtual: true,
	}
)

func TestGetVM(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	// vm exists
	vm, err := client.GetVM(1)
	assert.NoError(t, err)
	require.NotEmpty(t, vm)
	assert.Equal(t, vmA, vm)

	// vm is missing
	vm, err = client.GetVM(99999)
	assert.NoError(t, err)
	assert.Empty(t, vm)
}

func TestGetVMs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	vms, err := client.GetVMs()
	assert.NoError(t, err)
	require.NotEmpty(t, vms)
	assert.Equal(t, []*Device{vmA, vmB, vmC}, vms)
}

func TestGetVMsByTag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	vms, err := client.GetVMsByTag("node_exporter")
	assert.NoError(t, err)
	require.Len(t, vms, 2)

	// validating contents
	assert.Equal(t, []*Device{vmA, vmB}, vms)

	// tag doesn't exist
	vms, err = client.GetVMsByTag("doesn_t-exist")
	assert.NoError(t, err)
	assert.Empty(t, vms)
}
