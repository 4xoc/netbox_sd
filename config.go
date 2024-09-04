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
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

// Config is a generic config struct for netbox_sd
type Config struct {
	BaseURL            string        `yaml:"base_url"`
	Token              string        `yaml:"api_token"`
	AllowInsecure      bool          `yaml:"allow_insecure"`
	ScanIntervalString string        `yaml:"scan_interval"`
	ScanInterval       time.Duration `yaml:"-"`
	Groups             []*Group      `yaml:"groups"`
}

// Group contains specific configuration for groups to get targets for
type Group struct {
	File               string         `yaml:"file"`
	Type               string         `yaml:"type"`
	Match              string         `yaml:"match"`
	ScanIntervalString string         `yaml:"scan_interval"`
	ScanInterval       time.Duration  `yaml:"-"`
	Labels             model.LabelSet `yaml:"labels"`
	Port               *int           `yaml:"port"`
	Flags              Flags          `yaml:"flags"`
	Filters            []*Filter      `yaml:"filters"`
}

// Flags defines specific behavior that can be toggled on or off
type Flags struct {
	// IncludeVMs will cause VMs to be checked for matches too.
	IncludeVMs *bool `yaml:"include_vms"`
	// InetFamily defines which inet address family is returned. If an address of a target doesn't match the family, the
	// device is skipped in the resulting target group.
	InetFamily *string `yaml:"inet_family"`
	// AllAddresses causes all addresses of a service, device or interface to be returned when set to true. This still
	// honors the InetFamily filter.
	AllAddresses *bool `yaml:"all_addresses"`
}

// Filter defines a new filter where a the string index of the map is a label name and the value at that index
// represents a regular expression that must match.
type Filter struct {
	Label  string         `yaml:"label"`
	Match  string         `yaml:"match"`
	Negate bool           `yaml:"negate"`
	regex  *regexp.Regexp `yaml:"-"`
}

const (
	GroupTypeDeviceTag    = "device_tag"
	GroupTypeInterfaceTag = "interface_tag"
	GroupTypeService      = "service"
	InetFamilyAny         = "any"
	InetFamilyInet        = "inet"
	InetFamilyInet6       = "inet6"
)

var (
	ErrorBadFilterLabel    = errors.New("bad label for filter provided (must start with 'netbox_')")
	ErrorBadFilterMatch    = errors.New("bad filter match provided")
	ErrorBadGroupType      = errors.New("bad group type value")
	ErrorBadInetFamily     = errors.New("bad inet_family value provided")
	ErrorBadPort           = errors.New("bad port value")
	ErrorBadScanInterval   = errors.New("failed to parse scan_interval")
	ErrorBaseURLMissingTLS = errors.New("netbox_base_url must start with https and support tls")
	ErrorDuplicateFile     = errors.New("duplicate file name in configuration")
	ErrorMissingFile       = errors.New("missing config file path")
	ErrorMissingRequired   = errors.New("missing one or more required config values")
	ErrorParsingFile       = errors.New("failed to parse config file")
	ErrorReadingFile       = errors.New("failed to read config file")
)

// ReadConfigFile reads and parses a given config file
func readConfigFile(file string) (*Config, error) {
	var (
		err         error
		fileContent []byte
		config      Config
		group       *Group
		knownFiles  map[string]int = make(map[string]int)
		ok          bool
		i           int
	)

	if file == "" {
		return nil, ErrorMissingFile
	}

	fileContent, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorReadingFile, err.Error())
	}

	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return nil, fmt.Errorf("%w: %s", ErrorParsingFile, err.Error())
	}

	// check for required values
	if config.BaseURL == "" ||
		config.Token == "" ||
		config.ScanIntervalString == "" ||
		len(config.Groups) == 0 {
		return nil, fmt.Errorf("global configuration: %w", ErrorMissingRequired)
	}

	if !strings.HasPrefix(config.BaseURL, "https") {
		return nil, ErrorBaseURLMissingTLS
	}

	// parse scan_interval
	config.ScanInterval, err = time.ParseDuration(config.ScanIntervalString)
	if err != nil {
		return nil, ErrorBadScanInterval
	}

	// check all groups for required values & sanity
	for i, group = range config.Groups {
		// check for duplicate file name
		if _, ok = knownFiles[group.File]; ok {
			return nil, ErrorDuplicateFile
		} else {
			// add new file to knownFiles
			knownFiles[group.File] = 1
		}

		if err = validateGroup(group, &config); err != nil {
			return nil, fmt.Errorf("failed to validate group config with index %d: %w", i, err)
		}
	}

	return &config, nil
}

// ValidateGroup checks the contents of group.
func validateGroup(group *Group, config *Config) error {
	var (
		err error
	)

	if group.File == "" ||
		group.Type == "" ||
		group.Match == "" {
		return ErrorMissingRequired
	}

	if group.Type != GroupTypeService &&
		group.Type != GroupTypeDeviceTag &&
		group.Type != GroupTypeInterfaceTag {
		return ErrorBadGroupType
	}

	if group.ScanIntervalString != "" {
		// parse scan_interval
		group.ScanInterval, err = time.ParseDuration(group.ScanIntervalString)
		if err != nil {
			return ErrorBadScanInterval
		}
	} else {
		// use default
		group.ScanInterval = config.ScanInterval
	}

	if group.Port != nil {
		if *group.Port < 0 || *group.Port > 65535 {
			// port is invalid
			return ErrorBadPort
		}
	}

	// start checking flags
	if group.Flags.IncludeVMs == nil {
		// setting default
		group.Flags.IncludeVMs = new(bool)
		*group.Flags.IncludeVMs = true
	}

	if group.Flags.InetFamily == nil {
		// setting default
		group.Flags.InetFamily = new(string)
		*group.Flags.InetFamily = InetFamilyAny
	} else if *group.Flags.InetFamily != InetFamilyAny &&
		*group.Flags.InetFamily != InetFamilyInet &&
		*group.Flags.InetFamily != InetFamilyInet6 {

		return ErrorBadInetFamily
	}

	if group.Flags.AllAddresses == nil {
		// setting default
		group.Flags.AllAddresses = new(bool)
		*group.Flags.AllAddresses = false
	}

	return validateFilters(group.Filters)
}

// ValidateFilters checks that filters are valid.
func validateFilters(filters []*Filter) error {
	var (
		filter *Filter
		err    error
	)

	for _, filter = range filters {
		// Labels must start with `netbox_` to match in any case.
		if !strings.HasPrefix(filter.Label, "netbox_") {
			return ErrorBadFilterLabel
		}

		filter.regex, err = regexp.Compile(filter.Match)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrorBadFilterMatch, err.Error())
		}
	}

	return nil
}
