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
	"fmt"
	"net/netip"
	"regexp"
)

// Values of IP status as in IP.Status.Value
const (
	queryIPAddressAttributes string = "id address status vrf {id, name}"
	queryIPByAddress         string = "{ip_address_list(address:\"%s\"){" + queryIPAddressAttributes + "}}"
	queryInterfaceIPs        string = "{ip_address_list(interface_id:\"%d\"){" + queryIPAddressAttributes + "}}"
	queryVirtualInterfaceIPs string = "{ip_address_list(vminterface_id:\"%d\"){" + queryIPAddressAttributes + "}}"
)

var (
	cidrRegexp *regexp.Regexp = regexp.MustCompile(`(/\d{0,128})$`)
)

// IP describes a subset of details of a Netbox ip.
type IP struct {
	ID       uint64 `json:"-"`
	IDString string `json:"id"`
	Address  string `json:"address"`
	Status   string `json:"status"`
	VRF      *VRF   `json:"vrf"`
}

// Family returns the decimal number of the version that this IP represents.
func (ip *IP) Family() int {
	if netip.MustParseAddr(ip.ToAddr()).Is6() {
		return 6
	} else {
		return 4
	}
}

// GetIPsByAddress returns a list of netbox IP object based on a given address string (legacy IP or IPv6). This is the
// default option to get any IP object since an address can exist multiple times in various VRFs.  The caller is
// responslible to filter through the result to find the IP it's looking for.
func (client *Client) GetIPsByAddress(ip string) ([]*IP, error) {
	var (
		query   string = fmt.Sprintf(queryIPByAddress, ip)
		resp    response
		wrapper graphQLResponseWrapper
		err     error
	)

	resp, err = client.graphQL(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query api: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, ErrUnexpectedStatusCode
	}

	err = json.Unmarshal(resp.RawBody().Bytes(), &wrapper)
	if err != nil {
		client.promFailure.Inc()
		return nil, fmt.Errorf("failed to unmarshal json from response body buffer: %w", err)
	}

	if len(wrapper.Data.IPList) == 0 {
		// No matching IP was found.
		return nil, nil
	}

	// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
	wrapper.parseIDs()

	return wrapper.Data.IPList, nil
}

// GetInterfaceIPs returns a list of all IPs associated with a given dcim interface id.
func (client *Client) GetInterfaceIPs(id uint64) ([]*IP, error) {
	var (
		query   string = fmt.Sprintf(queryInterfaceIPs, id)
		resp    response
		wrapper graphQLResponseWrapper
		err     error
	)

	resp, err = client.graphQL(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query api: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, ErrUnexpectedStatusCode
	}

	err = json.Unmarshal(resp.RawBody().Bytes(), &wrapper)
	if err != nil {
		client.promFailure.Inc()
		return nil, fmt.Errorf("failed to unmarshal json from response body buffer: %w", err)
	}

	// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
	wrapper.parseIDs()

	return wrapper.Data.IPList, nil
}

// GetVirtualInterfaceIPs returns a list of all IPs associated with a given virtual interface id.
func (client *Client) GetVirtualInterfaceIPs(id uint64) ([]*IP, error) {
	var (
		query   string = fmt.Sprintf(queryVirtualInterfaceIPs, id)
		resp    response
		wrapper graphQLResponseWrapper
		err     error
	)

	resp, err = client.graphQL(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query api: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, ErrUnexpectedStatusCode
	}

	err = json.Unmarshal(resp.RawBody().Bytes(), &wrapper)
	if err != nil {
		client.promFailure.Inc()
		return nil, fmt.Errorf("failed to unmarshal json from response body buffer: %w", err)
	}

	// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
	wrapper.parseIDs()

	return wrapper.Data.IPList, nil
}

// ToAddr converts a given IP struct to a single IP (i.e. converting cidr to address).
func (ip *IP) ToAddr() string {
	return cidrRegexp.ReplaceAllString(ip.Address, "")
}
