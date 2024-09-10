// MIT License
//
// Copyright (c) 2024 WIIT AG
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit
// persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
// Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
// WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
// OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"fmt"
	"log"

	"github.com/4xoc/netbox_sd/internal/config"
	"github.com/4xoc/netbox_sd/pkg/netbox"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

// selectAddr takes a given list of netbox.IP and group config and checks which IPs should be included in the target's
// list. It filters by flags defined in the Group (like InetFamily and AllAddresses).
func selectAddr(addrs []*netbox.IP, group *config.Group) []*netbox.IP {
	var (
		firstInet6 *netbox.IP
		firstInet  *netbox.IP
		addr       *netbox.IP
		result     []*netbox.IP = make([]*netbox.IP, 0)
	)

	// Filtering all addrs by expected inetFamily and number of addrs (see flag AllAddresses)
	for _, addr = range addrs {

		// Calling functions typically don't check if an addr is not nil, so doing this here where it is more efficient. See
		// for example getTargetsByDeviceTag() in device.go.
		if addr == nil {
			continue
		}

		// Not all IPs are created equal.
		if addr.Status != netbox.StatusIPActive &&
			addr.Status != netbox.StatusIPDHCP &&
			addr.Status != netbox.StatusIPSLAAC {
			continue
		}

		switch addr.Family() {
		case 6:
			if *group.Flags.InetFamily == config.InetFamilyInet6 ||
				*group.Flags.InetFamily == config.InetFamilyAny {
				// Inet Family of addr matches flag filters.

				if firstInet6 == nil {
					// If only one addr is to be returned we want the first addr to be used.
					firstInet6 = addr
				}

				if *group.Flags.AllAddresses {
					// Adding all addrs.
					if !addrExists(addr, result) {
						result = append(result, addr)
					}
				} else {
					// Exit early from this loop when !AllAddresses because we've already got a good inet6 addr.
					break
				}
			}

		case 4:
			if *group.Flags.InetFamily == config.InetFamilyInet ||
				*group.Flags.InetFamily == config.InetFamilyAny {
				// Inet Family of addr matches flag filters.

				if firstInet == nil {
					// If only one addr is to be returned we want the first addr to be used.
					firstInet = addr
				}

				if *group.Flags.AllAddresses {
					// Adding all addrs.
					if !addrExists(addr, result) {
						result = append(result, addr)
					}
				}
			}

		default:
			log.Printf("got unsupported address family %d from netbox", addr.Family())
			return make([]*netbox.IP, 0)
		}
	}

	if len(result) == 0 {
		// If no result exists yet, first trying to add inet6 then if no v6 addr exists, trying to add legacy IP instead.
		// Otherwise no matching IP is returned *shrug*
		if firstInet6 != nil {
			result = append(result, firstInet6)
		} else if firstInet != nil {
			result = append(result, firstInet)
		}
	}

	return result
}

// AddrExists checks if a given netbox.IP is already existing in a []*netbox.IP
func addrExists(needle *netbox.IP, haystack []*netbox.IP) bool {
	var i int

	for i = range haystack {
		if haystack[i].Address == needle.Address {
			return true
		}
	}
	return false
}

// GenerateCustomFieldLabels generates based on a list of Netbox's custom fields an additional LabelSet. Should any of
// the custom fields fail to convert, an error is returned and the resulting labelSet should be ignored. All labels are
// prefixed with `netbox_`.
func generateCustomFieldLabels(cfm netbox.CustomFieldMap) (model.LabelSet, error) {
	var (
		allLabels model.LabelSet
		gotError  error
	)

	cfm.GetAllEntries(func(key string, val *netbox.CustomField) {
		var (
			label   model.LabelSet
			tmpStr  string
			tmpNum  float64
			tmpBool bool
			err     error
		)

		switch val.Datatype {
		case netbox.CustomFieldText:
			tmpStr, err = val.AsString()
			if err != nil {
				gotError = err
				log.Printf("failed to get custom field value as string: %v", err)
			}

			label = model.LabelSet{
				model.LabelName("netbox_" + key): model.LabelValue(tmpStr),
			}

		case netbox.CustomFieldNumber:
			tmpNum, err = val.AsFloat()
			if err != nil {
				gotError = err
				log.Printf("failed to get custom field value as float64: %v", err)
			}

			label = model.LabelSet{
				model.LabelName("netbox_" + key): model.LabelValue(fmt.Sprintf("%d", int64(tmpNum))),
			}

		case netbox.CustomFieldBool:
			tmpBool, err = val.AsBool()
			if err != nil {
				gotError = err
				log.Printf("failed to get custom field value as bool: %v", err)
			}

			label = model.LabelSet{
				model.LabelName("netbox_" + key): model.LabelValue(fmt.Sprintf("%t", tmpBool)),
			}

		}

		allLabels = allLabels.Merge(label)
	})

	// returns an error if any of the custom fields caused an error
	return allLabels, gotError
}

// SetTargetStatusMetric sets the PromTargetStatus metric for a given Device in group to state.
func SetTargetStatusMetric(group string, dev *netbox.Device, state TargetState) {
	promTargetState.
		With(prometheus.Labels{
			"group":                group,
			"netbox_name":          dev.Name,
			"netbox_rack":          dev.Rack.Name,
			"netbox_site":          dev.Site.Name,
			"netbox_tenant":        dev.Tenant.Name,
			"netbox_role":          dev.Role.Name,
			"netbox_serial_number": dev.SerialNumber,
			"netbox_asset_tag":     dev.AssetTag,
		}).Set(float64(state))
}

// ConvertToTargets takes a list of IPs and optional port and normalizes it into a slice of LabelSets.
func convertToTargets(ips []*netbox.IP, port *int) []model.LabelSet {
	var (
		// Init targets with appropriate capacity.
		targets = make([]model.LabelSet, 0, len(ips))
		i       int
	)

	for i = range ips {
		// Port is optional, thus only appending it when defined.
		if port != nil {
			if ips[i].Family() == 4 {
				targets = append(targets, model.LabelSet{
					model.AddressLabel: model.LabelValue(fmt.Sprintf("%s:%d", ips[i].ToAddr(), *port)),
				})
			} else {
				// IPv6 requires wrapping in brackets.
				targets = append(targets, model.LabelSet{
					model.AddressLabel: model.LabelValue(fmt.Sprintf("[%s]:%d", ips[i].ToAddr(), *port)),
				})
			}
		} else {
			targets = append(targets, model.LabelSet{
				model.AddressLabel: model.LabelValue(ips[i].ToAddr()),
			})
		}
	}

	return targets
}
