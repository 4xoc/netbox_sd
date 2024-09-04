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

// This file contains GraphQL specific functions.

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// GraphQLResponse is an implementation of the response interface.
type graphQLResponse struct {
	statusCode int
	body       bytes.Buffer
}

func (r *graphQLResponse) StatusCode() int {
	return r.statusCode
}

func (r *graphQLResponse) RawBody() *bytes.Buffer {
	return &r.body
}

// GraphQLResponseWrapper is a structure for extracting data from a GraphQL response body. A downstream function can use
// it to extract the parts of any GraphQL query it's interested in.
type graphQLResponseWrapper struct {
	Data struct {
		Device        *Device      `json:"device"`
		DeviceList    []*Device    `json:"device_list"`
		VM            *Device      `json:"virtual_machine"`
		VMList        []*Device    `json:"virtual_machine_list"`
		Interface     *Interface   `json:"interface"`
		InterfaceList []*Interface `json:"interface_list"`
		IP            *IP          `json:"ip_address"`
		IPList        []*IP        `json:"ip_address_list"`
		ServiceList   []*Service   `json:"service_list"`
	} `json:"data"`
}

// GraphQL performs a new GraphQL request towards Netbox, using query as GraphQL compliant query string. No validation
// of query is performed. No pagenation is used. On success a ptr to a Response struct is returned while error is not.
// The contents of the request is not further validated. Success therefore means some 2xx response code has been
// returned by Netbox. Otherwise error contains details about the failure and a nil ptr for Response is returned.
func (client *Client) graphQL(query string) (response, error) {
	var (
		resp        *http.Response
		gResp       graphQLResponse
		req         http.Request
		err         error
		dump, dump2 []byte
		body        string

		// used for request timing
		timer time.Time
		dur   time.Duration
	)

	body = "{\"query\":\"" + strings.ReplaceAll(query, "\"", "\\\"") + "\"}"

	req = http.Request{
		Method: http.MethodPost,
		Header: map[string][]string{
			"Accept":        {"application/json"},
			"Content-Type":  {"application/json"},
			"Authorization": {fmt.Sprintf("Token %s", client.token)},
		},
		Body: io.NopCloser(bytes.NewBufferString(body)),
		// sad panda - netbox-docker doesn't support chunked encoding
		ContentLength:    int64(len(body)),
		TransferEncoding: []string{"identity"},
	}

	req.URL, _ = url.ParseRequestURI(client.url + "/graphql/")

	timer = time.Now()
	resp, err = client.http.Do(&req)
	if err != nil {
		client.promError.
			With(prometheus.Labels{
				"url": "/graphql/",
			}).
			Inc()
		return nil, fmt.Errorf("http graphql call failed: %w", err)
	}

	defer resp.Body.Close()

	// calc request duration
	dur = time.Since(timer)

	client.promDuration.
		With(prometheus.Labels{
			"url":  "/graphql/",
			"code": strconv.Itoa(resp.StatusCode),
		}).
		Set(float64(dur * time.Nanosecond))

	client.promStatus.
		With(prometheus.Labels{
			"url":  "/graphql/",
			"code": strconv.Itoa(resp.StatusCode),
		}).
		Inc()

	// putting data into Response
	gResp.statusCode = resp.StatusCode
	_, err = gResp.body.ReadFrom(resp.Body)
	if err != nil {
		client.promFailure.Inc()
		return nil, fmt.Errorf("failed to read response body into buffer: %w", err)
	}

	if client.httpTracing {
		// It is more efficient to check the level instead of dumping the entire requests and response every time and just
		// throwing away the result.

		// Not enabling body dump because the io.Readers are empty at this point.
		dump, err = httputil.DumpRequest(&req, false)
		if err != nil {
			client.promFailure.Inc()
			client.log.Errorf("failed to dump http request: %v", err)
		} else {
			dump2, err = httputil.DumpResponse(resp, false)
			if err != nil {
				client.promFailure.Inc()
				client.log.Errorf("failed to dump http response: %v", err)
			} else {
				client.log.Tracef("===> HTTP Request <===\n%s%s\n\n", string(dump), body)
				client.log.Tracef("===> HTTP Response <===\n%s%s\n\n", string(dump2), gResp.body.String())
			}
		}
	}

	client.log.Tracef("http call took %dms", dur.Milliseconds())

	return &gResp, nil
}
