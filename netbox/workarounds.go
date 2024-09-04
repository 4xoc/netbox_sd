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
	"strconv"
)

// ID returns the nummeric id of a device. Used as helper function to generate correct ID.
//
// This is a workaround for broken graphql types being returned by Netbox. IDs are represented as strings instead of
// ids.
// TODO: remove once fixed in Netbox (https://github.com/netbox-community/netbox/issues/11472)
func parseNetboxID(idString string) uint64 {
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		panic("got id from netbox that couldn't be parsed to uint64")
	}
	return id
}

// ParseIDs checks for all possible IDs to be converted from string to uint64.
func (w *graphQLResponseWrapper) parseIDs() {
	if w.Data.Device != nil {
		w.Data.Device.parseIDs()
	}

	for i := range w.Data.DeviceList {
		w.Data.DeviceList[i].parseIDs()
	}

	if w.Data.VM != nil {
		w.Data.VM.parseIDs()
	}

	for i := range w.Data.VMList {
		w.Data.VMList[i].parseIDs()
	}

	if w.Data.Interface != nil {
		w.Data.Interface.parseIDs()
	}

	for i := range w.Data.InterfaceList {
		w.Data.InterfaceList[i].parseIDs()
	}

	if w.Data.IP != nil {
		w.Data.IP.parseIDs()
	}

	for i := range w.Data.IPList {
		w.Data.IPList[i].parseIDs()
	}

	for i := range w.Data.ServiceList {
		w.Data.ServiceList[i].parseIDs()
	}
}

func (d *Device) parseIDs() {
	d.ID = parseNetboxID(d.IDString)

	if d.PrimaryIP6 != nil {
		d.PrimaryIP6.ID = parseNetboxID(d.PrimaryIP6.IDString)
	}

	if d.PrimaryIP4 != nil {
		d.PrimaryIP4.ID = parseNetboxID(d.PrimaryIP4.IDString)
	}
}

func (i *Interface) parseIDs() {
	i.ID = parseNetboxID(i.IDString)

	if i.Device != nil {
		i.Device.parseIDs()
	}
}

func (ip *IP) parseIDs() {
	ip.ID = parseNetboxID(ip.IDString)
	if ip.VRF != nil {
		// vrf can be nil when the IP is in `global`
		ip.VRF.ID = parseNetboxID(ip.VRF.IDString)
	}
}

func (s *Service) parseIDs() {
	s.ID = parseNetboxID(s.IDString)

	if s.Device != nil {
		s.Device.parseIDs()
	}

	if s.VM != nil {
		s.VM.parseIDs()
	}

	for i := range s.IPAddresses {
		s.IPAddresses[i].parseIDs()
	}
}
