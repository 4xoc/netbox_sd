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
)

const (
	queryVMAttributes string = "id name primary_ip4{" + queryIPAddressAttributes + "} primary_ip6{" + queryIPAddressAttributes + "} custom_fields site{name} tenant{name} platform{name} role{name} status tags{name}"
	queryVM           string = "{virtual_machine(id:%d){" + queryVMAttributes + "}}"
	queryVMs          string = "{virtual_machine_list{" + queryVMAttributes + "}}"
	queryVMsByTag     string = "{virtual_machine_list(tag:\"%s\"){" + queryVMAttributes + "}}"
)

// IsVirtual returns true if the device represents a virtual machine.
func (d *Device) IsVirtual() bool {
	return d.isVirtual
}

// GetVM returns information about a VM gathered from Netbox. When error is not nil, the request failed and error gives
// further details what went wrong. VM might point to an invalid address at that point and must not be used whenever an
// error has been returned. When no vm with the given ID has been found, Device as well as error are nil.
func (client *Client) GetVM(id uint64) (*Device, error) {
	var (
		query   string = fmt.Sprintf(queryVM, id)
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

	if wrapper.Data.VM == nil {
		return nil, nil
	}

	wrapper.Data.VM.isVirtual = true

	// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
	wrapper.parseIDs()

	return wrapper.Data.VM, nil
}

// GetVMs returns a list of all VMs.
func (client *Client) GetVMs() ([]*Device, error) {
	var (
		err     error
		resp    response
		wrapper graphQLResponseWrapper
	)

	resp, err = client.graphQL(queryVMs)
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

	for i := range wrapper.Data.VMList {
		wrapper.Data.VMList[i].isVirtual = true

		// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
		wrapper.Data.VMList[i].parseIDs()
	}

	return wrapper.Data.VMList, nil
}

// GetVMsByTag returns a list of all vms with a given tag.
func (client *Client) GetVMsByTag(tag string) ([]*Device, error) {
	var (
		query   string = fmt.Sprintf(queryVMsByTag, tag)
		err     error
		resp    response
		wrapper graphQLResponseWrapper
		i       int
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

	for i = range wrapper.Data.VMList {
		wrapper.Data.VMList[i].isVirtual = true

		// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
		wrapper.Data.VMList[i].parseIDs()
	}

	return wrapper.Data.VMList, nil
}
