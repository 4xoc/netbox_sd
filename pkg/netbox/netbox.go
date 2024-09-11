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

// Package netbox provides types and methods to read Netbox objects via REST/GraphQL API. As of this writing, there are
// only a subset of objects and actions that are supported.
//
// This package implements the Prometheus collector interface to provide information about inner workings. For
// normalization of metrics all new instances of Client can be created with a given namespace (see Prometheus Go library
// for details) to attach to the existing namespace of your application. The subsystem name is always `netbox` and
// cannot be changed.
//
// Exported metrics:
//   - <namespace>_netbox_status{code,url} # number of API calls by response code and relative url
//   - <namespace>_netbox_error{url} # number of failed HTTP requests (due to network or whatever)
//   - <namespace>_netbox_failure # number of function invocations that resulted in an error being returned
//   - <namespace>_netbox_duration{code,url} # (last) duration it took to perform an HTTP request to Netbox by response code and url
//
// TODO: the logging stuff is probably wrong now
// By default this package logs through the Golang standard library log package. This is obviously annoying when adding
// this package to another project that uses it's own logging library. To easily change the log facility, this package
// provides special Interface called `Logger` that can be implemented by other packages and registered with
// `SetLogger()`. All log messages are then sent through the interface to the target functions that can then decide on
// how to process the message further.
//
// WARNING: Most Netbox objects in this library contain a public struct attribute `IDString`. This is only used for a
// workaround when handling GraphQL requests. DO NOT use this attribute anywhere as it will be removed once the bug has
// been fixed.
package netbox

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// SubsystemName is used for Prometheus metrics subsystem name.
const SubsystemName string = "netbox_api"

// Errors exported by this package.
var (
	ErrMissingURL           = errors.New("netbox url has not been provided")
	ErrMissingToken         = errors.New("netbox token has not been provided")
	ErrInvalidToken         = errors.New("provided token invalid or missing permissions")
	ErrInvalidURL           = errors.New("provided url invalid")
	ErrUnexpectedStatusCode = errors.New("received unexpected status code from netbox")
	ErrAmbiguous            = errors.New("provided search returned more than one possible result in netbox")
)

// defaultLog is an instance of defaultLogger used by this package.
var defaultLog defaultLogger

// Client describes a Netbox API client to perform REST calls with.
type Client struct {
	// URL contains the complete path to the base of Netbox's API (i.e. https://[..])
	url string
	// Token used for Netbox API queries.
	token string
	// HTTP client used across this instance
	http *http.Client

	// Logging options.
	log         Logger
	httpTracing bool // log http requests and resposes

	// Prometheus metrics for this instance.
	promNamespace string
	promStatus    *prometheus.CounterVec
	promError     *prometheus.CounterVec
	promFailure   prometheus.Counter
	promDuration  *prometheus.GaugeVec
}

// Value is a generic structure that is often used to define a label and value of some kind (think interface type, etc)
type Value struct {
	Value string `json:"value"`
}

// Name is a generic structure that is often used to define things like site, rack, etc.
type Name struct {
	Name string `json:"name"`
}

// NetboxStatus contains details about a Netbox installation.
type netboxStatus struct {
	Version string `json:"netbox-version"`
}

// New creates a new Client to interact with a netbox API. baseURL must point to a valid Netbox installation (without
// /api or /graphql at the end) while token must be a valid Netbox API key. WithTLS enabled TLS for HTTP transport while
// tlsInsecure can be set to allow any certificate to be accepted.
//
// In standard operation TLS should be used. System wide CAs are trusted.
func New(baseURL, token, promNamespace string, withTLS bool, tlsInsecure bool) (*Client, error) {
	var (
		client Client
		err    error
	)

	client.log = defaultLog
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)

	if token == "" {
		return nil, ErrMissingToken
	}

	if baseURL == "" {
		return nil, ErrMissingURL
	}

	_, err = url.Parse(baseURL)
	if err != nil {
		client.log.Errorf("given url could not be parsed: %v", err)
		return nil, ErrInvalidURL
	}

	client.url = baseURL
	client.token = token
	if withTLS {
		client.http = &http.Client{
			Transport: &http.Transport{
				// Set tls certificate validation.
				TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsInsecure},
			},
		}
	} else {
		client.http = http.DefaultClient
	}

	// Init Prometheus metrics
	client.promNamespace = promNamespace
	client.promStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   promNamespace,
			Subsystem:   SubsystemName,
			Name:        "status",
			Help:        "number of API calls",
			ConstLabels: nil,
		},
		[]string{"code", "url"},
	)

	client.promError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   promNamespace,
			Subsystem:   SubsystemName,
			Name:        "error",
			Help:        "number of http calls not completed due to errors",
			ConstLabels: nil,
		},
		[]string{"url"},
	)

	client.promFailure = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace:   promNamespace,
			Subsystem:   SubsystemName,
			Name:        "failure",
			Help:        "number of unexpected errors",
			ConstLabels: nil,
		})

	client.promDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   promNamespace,
			Subsystem:   SubsystemName,
			Name:        "duration_nanoseconds",
			Help:        "duration of api call",
			ConstLabels: nil,
		},
		[]string{"code", "url"},
	)

	return &client, nil
}

// VerifyConnectivity checks connectivity towards the netbox target machine. It also checks for validity of the API
// token. If connection and token are okay, nil is returned.
func (client *Client) VerifyConnectivity() error {
	var (
		resp   response
		err    error
		status netboxStatus
	)

	resp, err = client.get("/api/status/")
	if err != nil {
		return fmt.Errorf("failed to query api: %w", err)
	}

	// NOTE: It's not entirely safe to say that a 403 is sufficient here thus considering every non 200 code as bad token.
	if resp.StatusCode() != 200 {
		client.promFailure.Inc()
		return ErrInvalidToken
	}

	err = json.Unmarshal(resp.RawBody().Bytes(), &status)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json from response body buffer: %w", err)
	}

	if !netboxIsCompatible(status.Version) {
		return fmt.Errorf("detected incompatible Netbox version: v%s", status.Version)
	}

	return nil
}

// Copy creates and returns an identical copy of client. The http.Client is not duplicated but instead points to the
// same http.Client used for other copies. "[..] Clients should be reused instead of created as needed [..]" as per
// net/http docs.
func (client *Client) Copy() ClientIface {
	// TODO: needs prometheus stuff
	return &Client{
		url:         client.url,
		token:       client.token,
		http:        client.http,
		log:         client.log,
		httpTracing: client.httpTracing,
	}
}

// Describe implements the prometheus.Describe interface.
func (client *Client) Describe(ch chan<- *prometheus.Desc) {
	client.promStatus.Describe(ch)
	client.promError.Describe(ch)
	client.promDuration.Describe(ch)
	ch <- client.promFailure.Desc()
}

// Collect implements the prometheus.Collect interface.
func (client *Client) Collect(ch chan<- prometheus.Metric) {
	client.promStatus.Collect(ch)
	client.promError.Collect(ch)
	client.promDuration.Collect(ch)
	ch <- client.promFailure
}
