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

	"github.com/4xoc/netbox_sd/netbox"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// GetTargetsByDeviceTag returns a list of of target devices that match a given device tag.
func (sd *netboxSD) getTargetsByDeviceTag(group *Group) ([]*targetgroup.Group, error) {
	var (
		err         error
		dev         *netbox.Device
		dynLabels   model.LabelSet
		data        []*targetgroup.Group = make([]*targetgroup.Group, 0)
		target      *targetgroup.Group
		selectedIPs []*netbox.IP
		devList     []*netbox.Device
		vmList      []*netbox.Device
		cfLabels    model.LabelSet
	)

	devList, err = sd.api.GetDevicesByTag(group.Match)
	if err != nil {
		log.Printf("failed to get devices by tag")
		return nil, err
	}

	// Adding VMs with that tag here when flags are properly set.
	if *group.Flags.IncludeVMs {
		vmList, err = sd.api.GetVMsByTag(group.Match)
		if err != nil {
			log.Printf("failed to get vms by tag")
			return nil, err
		}

		devList = append(devList, vmList...)
	}

	for _, dev = range devList {

		// reset
		target = new(targetgroup.Group)

		// check for active device
		if dev.Status != netbox.StatusDeviceActive {
			log.Printf("device %s is not marked as active...skipping device", dev.Name)
			SetTargetStatusMetric(group.File, dev, TargetSkippedBadStatus)
			continue
		}

		target.Labels = model.LabelSet{
			model.LabelName("netbox_name"):          model.LabelValue(dev.Name),
			model.LabelName("netbox_rack"):          model.LabelValue(dev.Rack.Name),
			model.LabelName("netbox_site"):          model.LabelValue(dev.Site.Name),
			model.LabelName("netbox_tenant"):        model.LabelValue(dev.Tenant.Name),
			model.LabelName("netbox_role"):          model.LabelValue(dev.Role.Name),
			model.LabelName("netbox_platform"):      model.LabelValue(dev.Platform.Name),
			model.LabelName("netbox_serial_number"): model.LabelValue(dev.SerialNumber),
			model.LabelName("netbox_asset_tag"):     model.LabelValue(dev.AssetTag),
		}

		// custom fields
		cfLabels, err = generateCustomFieldLabels(dev.CustomFields)
		if err != nil {
			log.Printf("failed to parse custom fields for device %s...skipping device", dev.Name)
			SetTargetStatusMetric(group.File, dev, TargetSkippedBadCustomField)
			continue
		}

		target.Labels = target.Labels.Merge(cfLabels)

		if dev.IsVirtual() {
			dynLabels = model.LabelSet{
				model.LabelName("is_vm"): model.LabelValue("true"),
			}
		}

		target.Labels = target.Labels.Merge(dynLabels)
		target.Source = "netbox_sd"

		// add additional labels
		target.Labels = target.Labels.Merge(group.Labels)

		if !group.filtersMatch(target) {
			log.Printf("device %s doesn't match applied filters...skipping device", dev.Name)
			SetTargetStatusMetric(group.File, dev, TargetSkippedNotMatchingFilters)
			continue
		}

		// Only possible IPs for a device tag target can be primary v6 or legacy ip.
		selectedIPs = selectAddr([]*netbox.IP{dev.PrimaryIP6, dev.PrimaryIP4}, group)

		// When there are no selectedIPs this target cannot be used.
		if len(selectedIPs) == 0 {
			SetTargetStatusMetric(group.File, dev, TargetSkippedNoValidIP)
			continue
		}

		target.Targets = convertToTargets(selectedIPs, group.Port)

		SetTargetStatusMetric(group.File, dev, TargetActive)

		// add target to list
		data = append(data, target)

		// set prom metric
		promIPSkipped.
			With(prometheus.Labels{
				"group":       group.File,
				"netbox_name": dev.Name,
			}).Set(float64(len([]*netbox.IP{dev.PrimaryIP6, dev.PrimaryIP4}) - len(selectedIPs)))
	}

	return data, nil
}