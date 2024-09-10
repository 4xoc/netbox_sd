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
	queryInterfaceAttributes        string = "id name description enabled mark_connected mgmt_only type mtu parent{id} lag{id} mode custom_fields device {" + queryDeviceAttributes + "} tags{name}"
	queryVirtualInterfaceAttributes string = "id name description enabled mtu parent{id} mode custom_fields device: virtual_machine{" + queryVMAttributes + "} tags{name}"
	queryInterface                  string = "{interface(id:%d){" + queryInterfaceAttributes + "}}"
	queryVirtualInterface           string = "{interface: vm_interface(id:%d){" + queryVirtualInterfaceAttributes + "}}"
	queryInterfacesByTag            string = "{interface_list(tag:\"%s\"){" + queryInterfaceAttributes + "}}"
	queryVirtualInterfacesByTag     string = "{interface_list: vm_interface_list(tag:\"%s\"){" + queryVirtualInterfaceAttributes + "}}"
)

// Interface describes a subset of details about a Netbox interface.
type Interface struct {
	ID           uint64  `json:"-"`
	IDString     string  `json:"id"`
	Name         string  `json:"name"`
	Enabled      bool    `json:"enabled"`
	CustomFields CFMap   `json:"custom_fields"`
	Device       *Device `json:"device"`
	Tags         []Name  `json:"tags"`
	isVirtual    bool    `json:"-"`
}

// GetInterface returns the device interface identified by id.
func (client *Client) GetInterface(id uint64) (*Interface, error) {
	var (
		query   string = fmt.Sprintf(queryInterface, id)
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

	return wrapper.Data.Interface, nil
}

// GetVirtualInterface returns the virtual interface identified by id.
func (client *Client) GetVirtualInterface(id uint64) (*Interface, error) {
	var (
		query   string = fmt.Sprintf(queryVirtualInterface, id)
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

	if wrapper.Data.Interface == nil {
		return nil, nil
	}

	wrapper.Data.Interface.isVirtual = true

	if wrapper.Data.Interface.Device != nil {
		wrapper.Data.Interface.Device.isVirtual = true
	}

	// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
	wrapper.parseIDs()

	return wrapper.Data.Interface, nil
}

// GetInterfacesByTag returns a list of all device interfaces having a specific tag set in Netbox.
func (client *Client) GetInterfacesByTag(tag string) ([]*Interface, error) {
	var (
		query   string = fmt.Sprintf(queryInterfacesByTag, tag)
		err     error
		resp    response
		wrapper graphQLResponseWrapper
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

	return wrapper.Data.InterfaceList, nil
}

// GetVirtualInterfacesByTag returns a list of all virtual interfaces having a specific tag set in Netbox.
func (client *Client) GetVirtualInterfacesByTag(tag string) ([]*Interface, error) {
	var (
		query   string = fmt.Sprintf(queryVirtualInterfacesByTag, tag)
		err     error
		resp    response
		wrapper graphQLResponseWrapper
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

	for i := range wrapper.Data.InterfaceList {
		wrapper.Data.InterfaceList[i].isVirtual = true

		if wrapper.Data.InterfaceList[i].Device != nil {
			wrapper.Data.InterfaceList[i].Device.isVirtual = true
		}

		// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
		wrapper.Data.InterfaceList[i].parseIDs()
	}

	return wrapper.Data.InterfaceList, nil
}
