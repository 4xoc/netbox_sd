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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	service1 = &Service{
		ID:          1,
		IDString:    "1",
		Name:        "SSH",
		Device:      devA,
		Ports:       []int{22},
		IPAddresses: []*IP{ip1},
		Protocol:    "TCP",
		CustomFields: CFMap{
			entries: map[string]*CustomField{
				"custom_field_N": &CustomField{
					Datatype: "text",
					Value:    "bla",
				},
			},
		},
	}
	service2 = &Service{
		ID:          2,
		IDString:    "2",
		Name:        "service-C",
		VM:          vmA,
		Ports:       []int{5353, 53},
		IPAddresses: []*IP{ip3, ip4},
		Protocol:    "UDP",
		CustomFields: CFMap{
			entries: map[string]*CustomField{
				"custom_field_N": &CustomField{
					Datatype: "text",
					Value:    "whatever",
				},
			},
		},
	}
	service3 = &Service{
		ID:          3,
		IDString:    "3",
		Name:        "service-B",
		VM:          vmB,
		Ports:       []int{9909},
		IPAddresses: []*IP{ip6},
		Protocol:    "SCTP",
		CustomFields: CFMap{
			entries: map[string]*CustomField{
				"custom_field_N": &CustomField{
					Datatype: "text",
					Value:    "this",
				},
			},
		},
	}
	service4 = &Service{
		ID:          4,
		IDString:    "4",
		Name:        "SSH",
		VM:          vmA,
		Ports:       []int{22},
		IPAddresses: []*IP{ip4},
		Protocol:    "TCP",
		CustomFields: CFMap{
			entries: map[string]*CustomField{},
		},
	}
	service5 = &Service{
		ID:          5,
		IDString:    "5",
		Name:        "SSH",
		VM:          vmB,
		Ports:       []int{22},
		IPAddresses: []*IP{ip6},
		Protocol:    "TCP",
		CustomFields: CFMap{
			entries: map[string]*CustomField{},
		},
	}
)

func TestGetServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	client := newTestClient(t)

	srv, err := client.GetServices()
	require.NoError(t, err)
	require.NotEmpty(t, srv)
	sort.Slice(srv, func(i, j int) bool { return srv[i].ID < srv[j].ID })
	assert.Equal(t, []*Service{service1, service2, service3, service4, service5}, srv)
}

func TestGetServicesByName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	srv, err := client.GetServicesByName("SSH")
	require.NoError(t, err)
	require.NotEmpty(t, srv)
	assert.Equal(t, []*Service{service1, service4, service5}, srv)

	// checking for services that don't exist
	srv, err = client.GetServicesByName("does_not-exist")
	require.NoError(t, err)
	require.Empty(t, srv)
}
