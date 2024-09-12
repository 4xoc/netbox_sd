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
	queryDeviceAttributes string = "id name primary_ip4{" + queryIPAddressAttributes + "} primary_ip6{" + queryIPAddressAttributes + "} custom_fields rack{name} site{name} role{name} tenant{name} platform{name} serial asset_tag status tags{name}"
	queryDevice           string = "{device(id:%d){" + queryDeviceAttributes + "}}"
	queryDevices          string = "{device_list{" + queryDeviceAttributes + "}}"
	queryDevicesByTag     string = "{device_list(filters: {tag: \"%s\"}){" + queryDeviceAttributes + "}}"
)

// Device describes a subset of details of a Netbox device.
type Device struct {
	ID           uint64 `json:"-"`
	IDString     string `json:"id"`
	Name         string `json:"name"`
	PrimaryIP4   *IP    `json:"primary_ip4"`
	PrimaryIP6   *IP    `json:"primary_ip6"`
	CustomFields CFMap  `json:"custom_fields"`
	Rack         Name   `json:"rack"`
	Site         Name   `json:"site"`
	Role         Name   `json:"role"`
	Tenant       Name   `json:"tenant"`
	Platform     Name   `json:"platform"`
	SerialNumber string `json:"serial"`
	AssetTag     string `json:"asset_tag"`
	Status       string `json:"status"`
	Tags         []Name `json:"tags"`
	isVirtual    bool   `json:"-"`
}

// GetDevice returns information about a device gathered from Netbox. When error is not nil, the request failed and
// error gives further details what went wrong. Dev might point to an invalid address at that point and must not be used
// whenever an error has been returned. When no device with the given ID has been found, Device as well as error are nil.
func (client *Client) GetDevice(id uint64) (*Device, error) {
	var (
		query   string = fmt.Sprintf(queryDevice, id)
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

	return wrapper.Data.Device, nil
}

// GetDevices returns a list of all devices.
func (client *Client) GetDevices() ([]*Device, error) {
	var (
		err     error
		resp    response
		wrapper graphQLResponseWrapper
	)

	resp, err = client.graphQL(queryDevices)
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

	return wrapper.Data.DeviceList, nil
}

// GetDevicesByTag returns a list of all devices with a given tag.
func (client *Client) GetDevicesByTag(tag string) ([]*Device, error) {
	var (
		query   string = fmt.Sprintf(queryDevicesByTag, tag)
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

	return wrapper.Data.DeviceList, nil
}
