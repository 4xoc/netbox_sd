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

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// GetTargetsByService returns a list of of target devices that match a given service name
func (sd *netboxSD) getTargetsByService(group *config.Group) ([]*targetgroup.Group, error) {
	var (
		err         error
		i, j        int
		dev         *netbox.Device
		dynLabels   model.LabelSet
		data        []*targetgroup.Group = make([]*targetgroup.Group, 0)
		target      *targetgroup.Group
		selectedIPs []*netbox.IP
		serv        *netbox.Service
		servList    []*netbox.Service
		cfLabels    model.LabelSet
	)

	servList, err = sd.api.GetServicesByName(group.Match)
	if err != nil {
		log.Printf("failed to get services")
		return nil, err
	}

	for _, serv = range servList {
		// reset
		target = new(targetgroup.Group)

		// check if VM should be included
		if serv.VM != nil && !*group.Flags.IncludeVMs {
			continue
		}

		// Normalize to dev.
		if serv.VM != nil {
			dev = serv.VM
		} else {
			dev = serv.Device
		}

		// check for active device
		if dev.Status != netbox.StatusDeviceActive {
			log.Printf("device %s is not marked as active...skipping device", dev.Name)
			SetTargetStatusMetric(group.File, dev, TargetSkippedBadStatus)
			continue
		}

		target.Labels = model.LabelSet{
			model.LabelName("netbox_service"):       model.LabelValue(serv.Name),
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

		cfLabels, err = generateCustomFieldLabels(serv.CustomFields)
		if err != nil {
			log.Printf("failed to parse custom fields for service %s on device %s...skipping device", serv.Name, dev.Name)
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

		if !group.FiltersMatch(target) {
			log.Printf("device %s doesn't match applied filters...skipping device", dev.Name)
			SetTargetStatusMetric(group.File, dev, TargetSkippedNotMatchingFilters)
			continue
		}

		selectedIPs = selectAddr(serv.IPAddresses, group)

		// When there are no selectedIPs this target cannot be used.
		if len(selectedIPs) == 0 {
			SetTargetStatusMetric(group.File, dev, TargetSkippedNoValidIP)
			continue
		}

		// overwrite port if given in group config
		if group.Port != nil {
			serv.Ports = make([]int, 1)
			serv.Ports[0] = *group.Port
		}

		// Unless AllAddresses is set to true, only the first port is used
		// TODO: does this make sense??
		if !*group.Flags.AllAddresses && len(serv.Ports) > 1 {
			j = serv.Ports[0]
			serv.Ports = make([]int, 1)
			serv.Ports[0] = j
		}

		SetTargetStatusMetric(group.File, dev, TargetActive)

		for i = range selectedIPs {
			for j = range serv.Ports {
				// adding ports
				if selectedIPs[i].Family() == 4 {
					target.Targets = append(target.Targets, model.LabelSet{
						model.AddressLabel: model.LabelValue(fmt.Sprintf("%s:%d", selectedIPs[i].ToAddr(), serv.Ports[j])),
					})
				} else {
					target.Targets = append(target.Targets, model.LabelSet{
						model.AddressLabel: model.LabelValue(fmt.Sprintf("[%s]:%d", selectedIPs[i].ToAddr(), serv.Ports[j])),
					})
				}
			}
		}

		// add target to list
		data = append(data, target)
	}

	return data, nil
}
