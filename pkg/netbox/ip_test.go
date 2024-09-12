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
	ip1 = &IP{
		ID:       1,
		IDString: "1",
		Address:  "2001:db8::1/64",
		Status:   StatusIPActive,
		VRF:      nil,
	}
	ip2 = &IP{
		ID:       2,
		IDString: "2",
		Address:  "10.0.0.1/24",
		Status:   StatusIPActive,
		VRF:      nil,
	}
	ip3 = &IP{
		ID:       3,
		IDString: "3",
		Address:  "10.0.0.3/24",
		Status:   StatusIPActive,
		VRF:      nil,
	}
	ip4 = &IP{
		ID:       4,
		IDString: "4",
		Address:  "2001:db8::3/64",
		Status:   StatusIPActive,
		VRF:      nil,
	}
	ip5 = &IP{
		ID:       5,
		IDString: "5",
		Address:  "10.0.0.2/24",
		Status:   StatusIPActive,
		VRF:      nil,
	}
	ip6 = &IP{
		ID:       6,
		IDString: "6",
		Address:  "2001:db8::2/64",
		Status:   StatusIPActive,
		VRF:      nil,
	}
	ip7 = &IP{
		ID:       7,
		IDString: "7",
		Address:  "2001:db8::4/64",
		Status:   StatusIPReserved,
		VRF:      nil,
	}
	ip8 = &IP{
		ID:       8,
		IDString: "8",
		Address:  "2001:db8::4/64",
		Status:   StatusIPDeprecated,
		VRF: &VRF{
			ID:       1,
			IDString: "1",
			Name:     "vrf-A",
		},
	}
)

func TestFamily(t *testing.T) {
	assert.Equal(t, ip1.Family(), 6)
	assert.Equal(t, ip2.Family(), 4)
	assert.Equal(t, ip3.Family(), 4)
	assert.Equal(t, ip4.Family(), 6)
	assert.Equal(t, ip5.Family(), 4)
	assert.Equal(t, ip6.Family(), 6)
}

func TestToAddr(t *testing.T) {
	var (
		data = []struct {
			src      *IP
			expected string
		}{
			{ip1, "2001:db8::1"},
			{ip2, "10.0.0.1"},
			{ip3, "10.0.0.3"},
			{ip4, "2001:db8::3"},
			{ip5, "10.0.0.2"},
			{ip6, "2001:db8::2"},

			// special cases where no ip exists in netbox
			{&IP{Address: "2001:db8::1"}, "2001:db8::1"},
			{&IP{Address: "2001:db8::1/128"}, "2001:db8::1"},
			{&IP{Address: "10.0.0.1/8"}, "10.0.0.1"},
			{&IP{Address: "10.0.0.1/32"}, "10.0.0.1"},
			{&IP{Address: "10.0.0.1"}, "10.0.0.1"},
		}
		i int
	)

	for i = range data {
		assert.Equal(t, data[i].src.ToAddr(), data[i].expected)
	}
}

func TestGetIsPByAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	// simple IP without any shenanigans
	ip, err := client.GetIPsByAddress("2001:db8::1")
	require.NoError(t, err)
	require.NotEmpty(t, ip)
	assert.Equal(t, []*IP{ip1}, ip)

	// simple legacy IP
	ip, err = client.GetIPsByAddress("10.0.0.2")
	require.NoError(t, err)
	require.NotEmpty(t, ip)
	assert.Equal(t, []*IP{ip5}, ip)

	// checking for IP that doesn't exist
	ip, err = client.GetIPsByAddress("::1")
	require.NoError(t, err)
	require.Empty(t, ip)

	ip, err = client.GetIPsByAddress("127.0.0.1")
	require.NoError(t, err)
	require.Empty(t, ip)

	// check multiple VRFs
	ip, err = client.GetIPsByAddress("2001:db8::4")
	require.NoError(t, err)
	require.NotEmpty(t, ip)
	assert.Equal(t, []*IP{ip7, ip8}, ip)
}

func TestGetInterfaceIPs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	ips, err := client.GetInterfaceIPs(iface1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, ips)
	assert.Equal(t, []*IP{ip2, ip1}, ips)

	// checking for interface that doesn't exist
	ips, err = client.GetInterfaceIPs(9999)
	require.NoError(t, err)
	require.Empty(t, ips)
}

func TestGetVirtualInterfaceIPs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	client := newTestClient(t)

	ips, err := client.GetVirtualInterfaceIPs(vIface1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, ips)
	assert.Equal(t, []*IP{ip3, ip4}, ips)

	// checking for interface that doesn't exist
	ips, err = client.GetVirtualInterfaceIPs(9999)
	require.NoError(t, err)
	require.Empty(t, ips)
}
