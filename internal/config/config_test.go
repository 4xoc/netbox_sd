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

package config

import (
	"regexp"
	"testing"
	"time"

	"github.com/4xoc/netbox_sd/internal/util"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	var (
		err      error
		result   *Config
		expected *Config = &Config{
			BaseURL:            "https://netbox.domain.tld",
			Token:              "680000000000000000000000000000000000s038",
			ScanIntervalString: "5m",
			ScanInterval:       time.Duration(5 * time.Minute),
			Groups: []*Group{
				&Group{
					File:               "junos_exporter.prom",
					Type:               GroupTypeDeviceTag,
					Match:              "junos_exporter",
					Port:               util.NewPtr[int](1234),
					ScanIntervalString: "20s",
					ScanInterval:       time.Duration(20 * time.Second),
					Labels: model.LabelSet{
						"foo": "bar",
					},
					Flags: Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string](InetFamilyAny),
						AllAddresses: util.NewPtr[bool](false),
					},
				},
				&Group{
					File:               "ipmi_exporter.prom",
					Type:               GroupTypeInterfaceTag,
					Match:              "ipmi_exporter",
					Port:               util.NewPtr[int](1234),
					ScanIntervalString: "5m",
					ScanInterval:       time.Duration(5 * time.Minute),
					Labels: model.LabelSet{
						"foo": "bar",
					},
					Flags: Flags{
						IncludeVMs:   util.NewPtr[bool](true),
						InetFamily:   util.NewPtr[string](InetFamilyAny),
						AllAddresses: util.NewPtr[bool](false),
					},
				},
				&Group{
					File:         "junos2.prom",
					Type:         GroupTypeService,
					Match:        "junos_exporter",
					ScanInterval: time.Duration(5 * time.Minute),
					Labels: model.LabelSet{
						"foo": "bar",
					},
					Port: util.NewPtr[int](9100),
					Flags: Flags{
						IncludeVMs:   util.NewPtr[bool](false),
						InetFamily:   util.NewPtr[string](InetFamilyInet),
						AllAddresses: util.NewPtr[bool](true),
					},
				},
				&Group{
					File:         "junos3.prom",
					Type:         GroupTypeService,
					Match:        "junos_exporter",
					ScanInterval: time.Duration(5 * time.Minute),
					Labels: model.LabelSet{
						"foo": "bar",
					},
					Port: nil,
					Flags: Flags{
						IncludeVMs:   util.NewPtr[bool](false),
						InetFamily:   util.NewPtr[string](InetFamilyInet),
						AllAddresses: util.NewPtr[bool](true),
					},
					Filters: []*Filter{
						&Filter{
							Label: "netbox_foo",
							Match: "(bar|blub)",
							regex: regexp.MustCompile("(bar|blub)"),
						},
						&Filter{
							Label: "netbox_bar",
							Match: "something[0-9]+",
							regex: regexp.MustCompile("something[0-9]+"),
						},
					},
				},
			},
		}
	)

	// good config file
	result, err = ReadConfigFile("testdata/config/good.yml")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// covers all groups
	assert.Equal(t, expected, result)

	// missing path
	_, err = ReadConfigFile("")
	assert.ErrorIs(t, err, ErrorMissingFile)

	// file missing
	_, err = ReadConfigFile("testdata/config/foo")
	assert.ErrorIs(t, err, ErrorReadingFile)

	// malformed yaml
	_, err = ReadConfigFile("testdata/config/malformed.yml")
	assert.ErrorIs(t, err, ErrorParsingFile)

	// missing required
	_, err = ReadConfigFile("testdata/config/missingRequired.yml")
	assert.ErrorIs(t, err, ErrorMissingRequired)

	// missing required in group
	_, err = ReadConfigFile("testdata/config/missingRequiredInGroup.yml")
	assert.ErrorIs(t, err, ErrorMissingRequired)

	// bad group type
	_, err = ReadConfigFile("testdata/config/badGroupType.yml")
	assert.ErrorIs(t, err, ErrorBadGroupType)

	// bad default scan interval
	_, err = ReadConfigFile("testdata/config/badScanInterval.yml")
	assert.ErrorIs(t, err, ErrorBadScanInterval)

	// bad group scan interval
	_, err = ReadConfigFile("testdata/config/badScanInterval2.yml")
	assert.ErrorIs(t, err, ErrorBadScanInterval)

	// duplicate file
	_, err = ReadConfigFile("testdata/config/duplicateFile.yml")
	assert.ErrorIs(t, err, ErrorDuplicateFile)

	// bad port
	_, err = ReadConfigFile("testdata/config/badPort.yml")
	assert.ErrorIs(t, err, ErrorParsingFile)

	// bad port2
	_, err = ReadConfigFile("testdata/config/badPort2.yml")
	assert.ErrorIs(t, err, ErrorBadPort)

	// bad inet family
	_, err = ReadConfigFile("testdata/config/badInetFamily.yml")
	assert.ErrorIs(t, err, ErrorBadInetFamily)

	// bad filter label
	_, err = ReadConfigFile("testdata/config/badFilterLabel.yml")
	assert.ErrorIs(t, err, ErrorBadFilterLabel)

	// bad filter match
	_, err = ReadConfigFile("testdata/config/badFilterMatch.yml")
	assert.ErrorIs(t, err, ErrorBadFilterMatch)
}

func TestFiltersMatch(t *testing.T) {
	var (
		group = Group{
			Filters: []*Filter{
				&Filter{
					Label: "netbox_foo",
					Match: "bar",
				},
				&Filter{
					Label: "netbox_foo2",
					Match: "(foo|bar)",
				},
				&Filter{
					Label: "netbox_foo3",
					Match: "[0-9]+",
				},
				&Filter{
					Label:  "netbox_foo4",
					Match:  "bar",
					Negate: true,
				},
			},
		}
		data = []struct {
			target   *targetgroup.Group
			expected bool
		}{
			{
				// should work
				target: &targetgroup.Group{
					Labels: model.LabelSet{
						"netbox_foo":  "bar",
						"netbox_foo2": "foo",
						"netbox_foo3": "123",
						"netbox_foo4": "123",
					},
				},
				expected: true,
			},
			{
				// missing label defined in filters should fail
				target: &targetgroup.Group{
					Labels: model.LabelSet{
						"netbox_foo":  "bar",
						"netbox_foo2": "foo",
					},
				},
				expected: false,
			},
			{
				// netbox_foo3 should fail
				target: &targetgroup.Group{
					Labels: model.LabelSet{
						"netbox_foo":  "bar",
						"netbox_foo2": "foo",
						"netbox_foo3": "abc",
					},
				},
				expected: false,
			},
			{
				// all label values are wrong
				target: &targetgroup.Group{
					Labels: model.LabelSet{
						"netbox_foo":  "this",
						"netbox_foo2": "should",
						"netbox_foo3": "fail",
					},
				},
				expected: false,
			},
			{
				// negate match and thus return false
				target: &targetgroup.Group{
					Labels: model.LabelSet{
						"netbox_foo":  "bar",
						"netbox_foo2": "foo",
						"netbox_foo3": "123",
						"netbox_foo4": "bar",
					},
				},
				expected: false,
			},
		}
		i int
	)

	// Filters must compile
	require.NoError(t, validateFilters(group.Filters))

	for i = range data {
		assert.Equal(t, data[i].expected, group.FiltersMatch(data[i].target))
	}
}
