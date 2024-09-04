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
	queryServiceAttributes string = "id name device {" + queryDeviceAttributes + "} virtual_machine {" + queryVMAttributes + "} ports ipaddresses {" + queryIPAddressAttributes + "} protocol custom_fields"
	queryServicesByName    string = "{service_list(name:\"%s\"){" + queryServiceAttributes + "}}"
	queryServices          string = "{service_list{" + queryServiceAttributes + "}}"
)

// Service describes a subset of details of a netbox service
type Service struct {
	ID           uint64  `json:"-"`
	IDString     string  `json:"id"`
	Name         string  `json:"name"`
	Device       *Device `json:"device"`
	VM           *Device `json:"virtual_machine"`
	Ports        []int   `json:"ports"`
	IPAddresses  []*IP   `json:"ipaddresses"`
	Protocol     string  `json:"protocol"`
	CustomFields CFMap   `json:"custom_fields"`
}

// GetServices returns a list of all services that exists in Netbox.
func (client *Client) GetServices() ([]*Service, error) {
	var (
		resp    response
		wrapper graphQLResponseWrapper
		err     error
	)

	resp, err = client.graphQL(queryServices)
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

	for i := range wrapper.Data.ServiceList {
		if wrapper.Data.ServiceList[i].VM != nil {
			wrapper.Data.ServiceList[i].VM.isVirtual = true
		}
	}

	// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
	wrapper.parseIDs()

	return wrapper.Data.ServiceList, nil
}

// GetServicesByName returns a list of all services that exists in Netbox based on the service's name.
func (client *Client) GetServicesByName(name string) ([]*Service, error) {
	var (
		query   string = fmt.Sprintf(queryServicesByName, name)
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

	for i := range wrapper.Data.ServiceList {
		if wrapper.Data.ServiceList[i].VM != nil {
			wrapper.Data.ServiceList[i].VM.isVirtual = true
		}
	}

	// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
	wrapper.parseIDs()

	return wrapper.Data.ServiceList, nil
}
