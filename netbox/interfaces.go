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
	"bytes"

	"github.com/prometheus/client_golang/prometheus"
)

// ClientIface defines function for interacting with the Netbox API.
type ClientIface interface {
	prometheus.Collector

	// Get performs a simple GET request against Netbox's API. It takes a URL (either absolute, including protocol,
	// hostname, etc *OR* a relative to the API path like `/dcim/devices..`. When  the request was successful, the
	// response is returned and error is nil. If the request could not be performed for whatever reason, error is not nil
	// and response *must* not be used further.
	get(string) (response, error)

	// GraphQL performs a new GraphQL request towards Netbox, using a GraphQL compliant query string. No validation of
	// query is performed. No pagenation is used. On success a ptr to a Response struct is returned while error is not.
	// The contents of the request is not further validated. Success therefore means some 2xx response code has been
	// returned by Netbox. Otherwise error contains details about the failure and a nil ptr for Response is returned.
	graphQL(string) (response, error)

	/*
	 * devices
	 */

	// GetDevice queries Netbox for a specific device, identified by its nummeric ID. An error is returned when the API
	// call failed. *Device and error may be nil when a device of the ID doesn't exist.
	GetDevice(uint64) (*Device, error)

	// GetDevices returns a list of all devices.
	GetDevices() ([]*Device, error)

	// GetDevicesByTag returns a list of all devices with a given tag.
	GetDevicesByTag(string) ([]*Device, error)

	/*
	 * interfaces
	 */

	// GetInterface returns a single interface identified by id.
	GetInterface(uint64) (*Interface, error)
	// GetInterfacesByTag returns a list of all interfaces having a specific tag set in Netbox.
	GetInterfacesByTag(string) ([]*Interface, error)

	// GetVirtualInterface returns a single VM interface identified by id.
	GetVirtualInterface(uint64) (*Interface, error)
	// GetVirtualInterfacesByTag returns a list of all VM interfaces having a specific tag set in Netbox.
	GetVirtualInterfacesByTag(string) ([]*Interface, error)

	/*
	 * IP addresses
	 */

	// GetIPsByAddress searches Netbox for an IP object based on an address string given. Address MUST NOT be a cidr. An
	// error is returned when the API call failed. *IP and error may be nil when no ip matches the given address.
	GetIPsByAddress(string) ([]*IP, error)

	// GetInterfaceIPs returns a list of all IPs associated with a given interface id.
	GetInterfaceIPs(uint64) ([]*IP, error)
	// GetVirtualInterfaceIPs returns a list of all IPs associated with a given virtual interface id.
	GetVirtualInterfaceIPs(uint64) ([]*IP, error)

	/*
	 * services
	 */

	// GetServices returns a list of all services that exists in Netbox.
	GetServices() ([]*Service, error)

	// GetServicesByName returns a list of all services that exists in Netbox based on the service's name.
	GetServicesByName(string) ([]*Service, error)

	/*
	 * VMs
	 */

	// GetVM returns a device/vm identified by id.
	GetVM(uint64) (*Device, error)

	// GetVMs returns a list of all VMs.
	GetVMs() ([]*Device, error)

	// GetVMsByTag returns a list of all vms with a given tag.
	GetVMsByTag(string) ([]*Device, error)

	/*
	 * utilities
	 */

	// SetLogger updates the instance of ClientIface with a new Logger implementation.
	SetLogger(Logger)
	// HTTPTracing allows for enabling/disabling http request tracing.
	HTTPTracing(bool)
	// Copy creates an identical copy of the Netbox client.
	Copy() ClientIface
	// VerifyConnectivity tries to connect to the Netbox API, read data from it and checks if this was successful. It
	// tries to differentiate errors and return ErrInvalidToken when connectivity was okay but Netbox refused to comply
	// because the token is not valid (no such token, missing permissions, etc).
	VerifyConnectivity() error
}

// CustomFieldMap contains custom fields defined in Netbox associated with an entity (like device, interface, etc). It
// is used to access those custom fields.
type CustomFieldMap interface {
	// GetEntry returns a pointer to the CustomField identified by name. If no CustomField of that name exists, nil is
	// returned.
	GetEntry(string) *CustomField
	// GetAllEntries iterates over all CustomFields and calls the callback function with the field's name and a pointer to
	// the CustomField as arguments.
	GetAllEntries(func(string, *CustomField))
}

type response interface {
	StatusCode() int
	RawBody() *bytes.Buffer
}
