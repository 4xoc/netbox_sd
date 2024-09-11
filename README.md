# Netbox Service-Discovery for Prometheus
Netbox_SD is an opinionated approach to generates file-sd formatted YAML files used by Prometheus to detect targets and
devices to scrape. The source of those targets is Netbox, using various configurable types (namely tags & services).
This also adds additional labels to each target and metrics for monitoring.

## Features
- single binary to run besides Prometheus
- support for services, device/VM & interface tags
- support for skipping non-active devices
- support for filtering results by label value (using regular expressions)
- support for selecting inet family, single or all IPs for a target
- support for overwriting service port
- support for additional static labels
- IPv6 focused
- configurable scan interval for each configured group
- adds extensive labels for each target (like name, rack, site, role, etc) defined in Netbox
- adds custom fields as labels
- all-or-nothing changes (communication errors with Netbox don't result in incomplete file_sd updates)
- extensive metrics for configured groups (including individual target state) & communication with Netbox to monitor
	Netbox_SD itself

## About this project
The basis of this project is the code base used some years now at WIIT AG in production. With the open sourcing of the
base code, this project aims to maintain and expand its features. There is currently no fixed plan where the development
will head towards but everyone is invited to request features, open bug issues and help develop this project to cover
more bases. A couple of ideas I have in mind:

- have `staged` status already add devices but create a silence for the device; similarly have `failed` do the same for
	already existing devices
- allow for label dropping within netbox_sd to remove unwanted labels and reduce cardinality
- more TLS features like fixed CA cert for internal CAs; client certificate authentication towards Netbox
- how about a web gui to get more insights into netbox_sd?

## Supported Netbox Versions
< 4.0.0 (v4 has not been tested, but is a high prio topic)  
\>= 3.4.5 required (or else you'll have a bad time)

## Usage Considerations
This software is intended to run on the same machine that Prometheus runs on. As a user, you must ensure that the files
written by Netbox_SD are accessible to Prometheus. Netbox_SD doesn't delete any files so when a group is deconfigured
it's the users responsibility to clean up after. This is intentional to reduce risk of Netbox_SD deleting other file-sd
files that are configured by different means (like some orchestration software).

Netbox_SD comes with metrics itself. Make sure you monitor netbox_sd_api_error_count and netbox_sd_api_status for 403
and 5xx errors. Netbox_SD will not change a group unless *all* API calls to Netbox have succeeded. Again, this is to
prevent targets being removed when Netbox is down.

You probably want to monitor netbox_sd_target_state and netbox_sd_addresses_skipped too. Targets can be ignored for
various reasons (like device not being `active`) and some being ignored should be a hint for you that they might not be
configured correctly in Netbox. Especially netbox_sd_target_state gives details about the exact reason a target was
ignored:
* 1 = active and added to target list
* 0 = ignored for unspecified reason (catch-all)
* -1 = skipped because device status is not active
* -2 = skipped because custom fields couldn't be processed (check for misconfiguration in Netbox or open a ticket here)
* -3 = skipped because no valid IP could be selected for target (e.g. because flags specified a different inet version)
* -4 = skipped because not all filters matched for this device

When a file cannot be updated (i.e. written to disk) netbox_sd_update_error shows that. This is not good. You should fix
that asap.

By choice, Netbox_SD does not provide the HTTP SD (https://prometheus.io/docs/prometheus/latest/http_sd/) mechanism.

## Default Labels
A handful of labels are automatically set for each target:
* netbox_name
* netbox_rack
* netbox_site
* netbox_tenant
* netbox_role
* netbox_serial_number
* netbox_asset_tag

## Custom Fields as Prometheus Labels
Custom fields for devices are automatically added unless empty. The syntax is always `netbox_$CustomFieldName`. The
name of the custom field is not changed (note this refers to the actual name, not a Label by itself that can contain
spaces for human readable names of a label). Case sensitivy being removed by Prometheu's library is not a bug but a
feature.

## Tags as Prometheus Labels
Feature to add tags as labels is planned.

## Configuration
Netbox_SD is configured using a yaml formatted file and pointed to using the `-config.file` command line argument.

```
# required: base URL of the Netbox installation
base_url: https://netbox.domain.tld/

# required: API token with read permissions
api_token: 1234567890

# required: default scan interval
scan_interval: 10s

# optional: skip ssl verification
# insecure_skip_verify: true
groups:
    # required: file name to write targets into
  - file: junos_exporter.yml
    # optional: group specific scan interval
    scan_interval: 5m

    # required: type of attribute to check in Netbox (device_tag, interface_tag or service)
    type: device_tag

    # required: string to match the type (i.e. service name or tag)
    match: junos_exporter

    # optional: adds a port to the target address; will overwrite a service port (if defined) and used with service type
    # WARNING: 0 is considered a valid port and will cause service ports to be overwritten
    port: 9100

    # optional: map of additional tags to add to each target
    labels:
      foo: bar
			
		# optional: filter by label values
		filters:
      # required: label to match for (must begin with `netbox_`)
      - label: netbox_rack
      	# required: regular expression that must match for the above label for the target to be included.
      	# see https://github.com/google/re2/wiki/Syntax for details
      	match: 'rack[0-9]+'
      	# optionally negate a match (making a regex that would otherwise match be excluded from the results); useful
      	# where golang regex doesn't support negate natively.
      	negate: true

    # additional flags to change behaviour for a particular group
    flags:
      # include vms in results; otherwise only devices are returned; affects device_tag, interface_tag & service type
      # default: true
      include_vms: [ true | false ]

      # Inet family a matching IP must be part of to be returned as target. When the IP doesn't match, the device is
      # not returned as target. When using `any` and `all_addresses` is false, inet6 takes precedence over inet.
      # Affects device_tag, interface_tag & service type
      # default: any
      inet_family: [ inet | inet6 | any ]

      # When true all IPs matching are returned as individual targets. Affects interface_tag only when more than one
      # ip has been assigned. inet_family is still considered meaning addresses of a family excluded in inet_family
      # will not be returned. When false only the first IP is used. No affect on service type as this is always a
      # single IP.
      # default: false
      all_addresses: [ true | false ]

  - file: junos_exporter_slow.yml
    scan_interval: 5m
    type: device_tag
    match: junos_exporter_slow
```

### Supported Types
- device_tag: tag added on the device level
- interface_tag: tag added on an interface level
- service: service definition

### Filters
Additional filters can be applied to targets found through tags. Filters work on all labels applied by netbox_sd and are
regex matches. The list of filters within a group configuration are _always_ an AND combination of filters.

### Port Override
By default a tag based group will only return the address without any port information. Only service adds the port
automatically. To ensure a port for a specific group is given, the `port` config option can be set (it's ignored for
service based group types). When `port` is defined, it's value is appended to the address.

## Metrics
Netbox_sd exposes prometheus-style metrics at `/metrics` on the configured `--web.listen=` address. The following
metrics (additional to golang specific ones) are exposed:
- netbox_sd_target_state{netbox_name, netbox_site, netbox..} (see [Usage Considerations](#usage-considerations))
- netbox_sd_update_timestamp{group}
- netbox_sd_update_error{group}
- update_duration_nanoseconds{group}
- netbox_sd_target_count{group}
- netbox_sd_target_skipped{group}
- netbox_sd_addresses_skipped{group,netbox_name}
- netbox_sd_api_status (200, 403, etc)
- netbox_sd_api_duration_seconds

## Noteworthy Mention
Special thanks goes out to [WIIT AG](https://www.wiit.cloud/en/) for open sourcing netbox_sd and netbox-go. This tool
has been developed and used in production for multiple years now internally and WIIT AG was kind enough to release this
under the MIT license.
