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

// This file contains REST specific functions.

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// restResponse is an implementation of the response interface.
type restResponse struct {
	// StatusCode is the HTTP status code returned by the server.
	statusCode int
	// Body contains the read response body in a buffer that can be read multiple times.
	body bytes.Buffer
}

func (r *restResponse) StatusCode() int {
	return r.statusCode
}

func (r *restResponse) RawBody() *bytes.Buffer {
	return &r.body
}

// Get performs a new HTTP Get request for a given apiURL towards Netbox. Query must be a relative path to BaseURL. If
// successfull, a non-nil response interface is returned while error is nil. Otherwise error contains details about what
// went wrong. reponse must not be used when error is not nil.
//
// This implementation doesn't support paging by itself but a calling function can supply the limit and offset parameter
// itself within query to do it itself.
func (client *Client) get(query string) (response, error) {
	var (
		resp        *http.Response
		rResp       restResponse
		req         http.Request
		err         error
		dump, dump2 []byte

		// used for request timing
		timer time.Time
		dur   time.Duration
	)

	req = http.Request{
		Method: http.MethodGet,
		Header: map[string][]string{
			"Accept":        {"application/json"},
			"Authorization": {fmt.Sprintf("Token %s", client.token)},
		},
	}

	req.URL, _ = url.ParseRequestURI(client.url + query)

	timer = time.Now()
	resp, err = client.http.Do(&req)
	if err != nil {
		client.promError.
			With(prometheus.Labels{
				"url": query,
			}).
			Inc()
		return nil, fmt.Errorf("http api call failed: %w", err)
	}

	defer resp.Body.Close()

	// calc request duration
	dur = time.Since(timer)

	client.promDuration.
		With(prometheus.Labels{
			"url":  query,
			"code": strconv.Itoa(resp.StatusCode),
		}).
		Set(float64(dur * time.Nanosecond))

	client.promStatus.
		With(prometheus.Labels{
			"url":  query,
			"code": strconv.Itoa(resp.StatusCode),
		}).
		Inc()

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
				client.log.Tracef("===> HTTP Request <===\n%s\n", string(dump))
				client.log.Tracef("===> HTTP Response <===\n%s%s\n\n", string(dump2), rResp.body.String())
			}
		}
	}

	client.log.Tracef("http call took %dms", dur.Milliseconds())

	// putting data into response
	rResp.statusCode = resp.StatusCode
	_, err = rResp.body.ReadFrom(resp.Body)
	if err != nil {
		client.promFailure.Inc()
		return nil, fmt.Errorf("failed to read response body into buffer: %w", err)
	}

	return &rResp, nil
}
