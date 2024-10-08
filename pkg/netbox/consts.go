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

// These constants are used to match values returned by Netbox to static names in go. Since Netbox happens to change API
// details frequently to address new use cases, these constants aim to provide some consistency across applications.
const (
	StatusDeviceActive          string = "active"
	StatusDeviceOffline         string = "offline"
	StatusDevicePlanned         string = "planned"
	StatusDeviceStaged          string = "staged"
	StatusDeviceFailed          string = "failed"
	StatusDeviceDecommissioning string = "decommissioning"

	StatusIPActive     string = "active"
	StatusIPReserved   string = "reserved"
	StatusIPDeprecated string = "deprecated"
	StatusIPDHCP       string = "dhcp"
	StatusIPSLAAC      string = "slaac"

	ServiceProtocolTCP  string = "tcp"
	ServiceProtocolUDP  string = "udp"
	ServiceProtocolSCTP string = "sctp"

	// because:
	// >=4.0.10 - https://github.com/netbox-community/netbox/issues/16946
	compatibleNetboxVersion string = ">=4.0.10"
)
