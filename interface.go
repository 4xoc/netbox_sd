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
	"log"

	"github.com/4xoc/netbox_sd/internal/config"
	"github.com/4xoc/netbox_sd/pkg/netbox"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// GetTargetsByInterfaceTag returns a list of of target devices that match a given device tag.
func (sd *netboxSD) getTargetsByInterfaceTag(group *config.Group) ([]*targetgroup.Group, error) {
	var (
		err         error
		iface       *netbox.Interface
		addrs       []*netbox.IP
		dynLabels   model.LabelSet
		data        []*targetgroup.Group = make([]*targetgroup.Group, 0)
		target      *targetgroup.Group
		selectedIPs []*netbox.IP
		ifList      []*netbox.Interface
		vmList      []*netbox.Interface
		cfLabels    model.LabelSet
	)

	ifList, err = sd.api.GetInterfacesByTag(group.Match)
	if err != nil {
		log.Printf("failed to get interfaces by tag: %v", err)
		return nil, err
	}

	// Adding virtual interfaces with that tag here when flags are properly set.
	if *group.Flags.IncludeVMs {
		vmList, err = sd.api.GetVirtualInterfacesByTag(group.Match)
		if err != nil {
			log.Printf("failed to get virtual images by tag: %v", err)
			return nil, err
		}

		ifList = append(ifList, vmList...)
	}

	for _, iface = range ifList {
		// reset
		target = new(targetgroup.Group)

		// check for active device & interface
		if iface.Device.Status != netbox.StatusDeviceActive ||
			!iface.Enabled {
			log.Printf("device %s is not marked as active...skipping device", iface.Device.Name)
			SetTargetStatusMetric(group.File, iface.Device, TargetSkippedBadStatus)
			continue
		}

		target.Labels = model.LabelSet{
			model.LabelName("netbox_name"):          model.LabelValue(iface.Device.Name),
			model.LabelName("netbox_rack"):          model.LabelValue(iface.Device.Rack.Name),
			model.LabelName("netbox_site"):          model.LabelValue(iface.Device.Site.Name),
			model.LabelName("netbox_tenant"):        model.LabelValue(iface.Device.Tenant.Name),
			model.LabelName("netbox_role"):          model.LabelValue(iface.Device.Role.Name),
			model.LabelName("netbox_platform"):      model.LabelValue(iface.Device.Platform.Name),
			model.LabelName("netbox_serial_number"): model.LabelValue(iface.Device.SerialNumber),
			model.LabelName("netbox_asset_tag"):     model.LabelValue(iface.Device.AssetTag),
		}

		// custom fields
		cfLabels, err = generateCustomFieldLabels(iface.Device.CustomFields)
		if err != nil {
			log.Printf("failed to parse custom fields for device %s...skipping device", iface.Device.Name)
			SetTargetStatusMetric(group.File, iface.Device, TargetSkippedBadCustomField)
			continue
		}

		target.Labels = target.Labels.Merge(cfLabels)

		cfLabels, err = generateCustomFieldLabels(iface.CustomFields)
		if err != nil {
			log.Printf("failed to parse custom fields for interface %s on device %s...skipping device", iface.Name, iface.Device.Name)
			SetTargetStatusMetric(group.File, iface.Device, TargetSkippedBadCustomField)
			continue
		}

		target.Labels = target.Labels.Merge(cfLabels)

		if iface.Device.IsVirtual() {
			dynLabels = model.LabelSet{
				model.LabelName("is_vm"): model.LabelValue("true"),
			}
		}

		target.Labels = target.Labels.Merge(dynLabels)
		target.Source = "netbox_sd"

		// add additional labels
		target.Labels = target.Labels.Merge(group.Labels)

		if !group.FiltersMatch(target) {
			log.Printf("device %s doesn't match applied filters...skipping device", iface.Device.Name)
			continue
		}

		// Only possible IPs for a device tag target can be primary v6 or legacy ip.
		if iface.Device.IsVirtual() {
			addrs, err = sd.api.GetVirtualInterfaceIPs(iface.ID)
		} else {
			addrs, err = sd.api.GetInterfaceIPs(iface.ID)
		}

		if err != nil {
			log.Printf("failed to get interface IPs for %s on %s...skipping device", iface.Name, iface.Device.Name)
			SetTargetStatusMetric(group.File, iface.Device, TargetSkippedNoValidIP)
			continue
		}

		selectedIPs = selectAddr(addrs, group)

		// When there are no selectedIPs this target cannot be used.
		if len(selectedIPs) == 0 {
			SetTargetStatusMetric(group.File, iface.Device, TargetSkippedNoValidIP)
			continue
		}

		target.Targets = convertToTargets(selectedIPs, group.Port)

		SetTargetStatusMetric(group.File, iface.Device, TargetActive)

		// add target to list
		data = append(data, target)

		// set prom metric
		promIPSkipped.
			With(prometheus.Labels{
				"group":       group.File,
				"netbox_name": iface.Device.Name,
			}).Set(float64(len(addrs) - len(selectedIPs)))
	}

	return data, nil
}
