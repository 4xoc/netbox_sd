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
	"regexp"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
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
					Port:               newPtr[int](1234),
					ScanIntervalString: "20s",
					ScanInterval:       time.Duration(20 * time.Second),
					Labels: model.LabelSet{
						"foo": "bar",
					},
					Flags: Flags{
						IncludeVMs:   newPtr[bool](true),
						InetFamily:   newPtr[string](InetFamilyAny),
						AllAddresses: newPtr[bool](false),
					},
				},
				&Group{
					File:               "ipmi_exporter.prom",
					Type:               GroupTypeInterfaceTag,
					Match:              "ipmi_exporter",
					Port:               newPtr[int](1234),
					ScanIntervalString: "5m",
					ScanInterval:       time.Duration(5 * time.Minute),
					Labels: model.LabelSet{
						"foo": "bar",
					},
					Flags: Flags{
						IncludeVMs:   newPtr[bool](true),
						InetFamily:   newPtr[string](InetFamilyAny),
						AllAddresses: newPtr[bool](false),
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
					Port: newPtr[int](9100),
					Flags: Flags{
						IncludeVMs:   newPtr[bool](false),
						InetFamily:   newPtr[string](InetFamilyInet),
						AllAddresses: newPtr[bool](true),
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
						IncludeVMs:   newPtr[bool](false),
						InetFamily:   newPtr[string](InetFamilyInet),
						AllAddresses: newPtr[bool](true),
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
	result, err = readConfigFile("testdata/config/good.yml")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// covers all groups
	assert.Equal(t, expected, result)

	// missing path
	_, err = readConfigFile("")
	assert.ErrorIs(t, err, ErrorMissingFile)

	// file missing
	_, err = readConfigFile("testdata/config/foo")
	assert.ErrorIs(t, err, ErrorReadingFile)

	// malformed yaml
	_, err = readConfigFile("testdata/config/malformed.yml")
	assert.ErrorIs(t, err, ErrorParsingFile)

	// missing required
	_, err = readConfigFile("testdata/config/missingRequired.yml")
	assert.ErrorIs(t, err, ErrorMissingRequired)

	// missing required in group
	_, err = readConfigFile("testdata/config/missingRequiredInGroup.yml")
	assert.ErrorIs(t, err, ErrorMissingRequired)

	// bad group type
	_, err = readConfigFile("testdata/config/badGroupType.yml")
	assert.ErrorIs(t, err, ErrorBadGroupType)

	// bad default scan interval
	_, err = readConfigFile("testdata/config/badScanInterval.yml")
	assert.ErrorIs(t, err, ErrorBadScanInterval)

	// bad group scan interval
	_, err = readConfigFile("testdata/config/badScanInterval2.yml")
	assert.ErrorIs(t, err, ErrorBadScanInterval)

	// duplicate file
	_, err = readConfigFile("testdata/config/duplicateFile.yml")
	assert.ErrorIs(t, err, ErrorDuplicateFile)

	// bad port
	_, err = readConfigFile("testdata/config/badPort.yml")
	assert.ErrorIs(t, err, ErrorParsingFile)

	// bad port2
	_, err = readConfigFile("testdata/config/badPort2.yml")
	assert.ErrorIs(t, err, ErrorBadPort)

	// bad inet family
	_, err = readConfigFile("testdata/config/badInetFamily.yml")
	assert.ErrorIs(t, err, ErrorBadInetFamily)

	// bad filter label
	_, err = readConfigFile("testdata/config/badFilterLabel.yml")
	assert.ErrorIs(t, err, ErrorBadFilterLabel)

	// bad filter match
	_, err = readConfigFile("testdata/config/badFilterMatch.yml")
	assert.ErrorIs(t, err, ErrorBadFilterMatch)
}
