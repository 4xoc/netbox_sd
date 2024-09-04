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
	"io"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type TargetState float64

const (
	PrometheusNameSpace             string      = "netbox_sd"
	TargetActive                    TargetState = 1
	TargetSkippedOther              TargetState = 0
	TargetSkippedBadStatus          TargetState = -1
	TargetSkippedBadCustomField     TargetState = -2
	TargetSkippedNoValidIP          TargetState = -3
	TargetSkippedNotMatchingFilters TargetState = -4
)

var (
	promInfo *prometheus.CounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   PrometheusNameSpace,
			Subsystem:   "",
			Name:        "info",
			Help:        "build information",
			ConstLabels: nil,
		},
		[]string{"version", "commit", "build_date"},
	)

	promGroups prometheus.Gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   PrometheusNameSpace,
			Subsystem:   "",
			Name:        "group_count",
			Help:        "Number of configured groups",
			ConstLabels: nil,
		})

	promTargetState *prometheus.GaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   PrometheusNameSpace,
			Subsystem:   "",
			Name:        "target_state",
			Help:        "state of specific target (see docs)",
			ConstLabels: nil,
		},
		[]string{
			"group",
			"netbox_name",
			"netbox_rack",
			"netbox_site",
			"netbox_tenant",
			"netbox_role",
			"netbox_serial_number",
			"netbox_asset_tag",
		},
	)

	promUpdateTime *prometheus.GaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   PrometheusNameSpace,
			Subsystem:   "",
			Name:        "update_timestamp",
			Help:        "Time in seconds since epoch when last update was done (successful or failed)",
			ConstLabels: nil,
		},
		[]string{"group"},
	)

	promUpdateError *prometheus.CounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   PrometheusNameSpace,
			Subsystem:   "",
			Name:        "update_error",
			Help:        "Number of update errors since process start",
			ConstLabels: nil,
		},
		[]string{"group"},
	)

	promUpdateDuration *prometheus.GaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   PrometheusNameSpace,
			Subsystem:   "",
			Name:        "update_duration_nanoseconds",
			Help:        "Time in nanoseconds taken for last update",
			ConstLabels: nil,
		},
		[]string{"group"},
	)

	promTargetCount *prometheus.GaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   PrometheusNameSpace,
			Subsystem:   "",
			Name:        "target_count",
			Help:        "Number of targets detected",
			ConstLabels: nil,
		},
		[]string{"group"},
	)

	promIPSkipped *prometheus.GaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   PrometheusNameSpace,
			Subsystem:   "",
			Name:        "addresses_skipped",
			Help:        "Number of ip addresses skipped by device",
			ConstLabels: nil,
		},
		[]string{"group", "netbox_name"},
	)
)

// Describe implements the prometheus.Describe interface.
func (sd *netboxSD) Describe(ch chan<- *prometheus.Desc) {
	ch <- promGroups.Desc()
	promInfo.Describe(ch)
	promUpdateTime.Describe(ch)
	promUpdateError.Describe(ch)
	promUpdateDuration.Describe(ch)
	promTargetCount.Describe(ch)
	promIPSkipped.Describe(ch)
	promTargetState.Describe(ch)

	if sd.api != nil {
		// Get metrics from netbox-go, when already initialized.
		sd.api.Describe(ch)
	}
}

// Collect implements the prometheus.Collect interface.
func (sd *netboxSD) Collect(ch chan<- prometheus.Metric) {
	ch <- promGroups
	promInfo.Collect(ch)
	promUpdateTime.Collect(ch)
	promUpdateError.Collect(ch)
	promUpdateDuration.Collect(ch)
	promTargetCount.Collect(ch)
	promIPSkipped.Collect(ch)
	promTargetState.Collect(ch)

	if sd.api != nil {
		// Get metrics from netbox-go, when already initialized.
		sd.api.Collect(ch)
	}
}

// serveMetrics starts an http server 
func (sd *netboxSD) serveMetrics(addr *string) {

	prometheus.MustRegister(sd)

	// Set promInfo only once.
	promInfo.With(prometheus.Labels{
		"version":    version,
		"commit":     commit,
		"build_date": date,
	}).Inc()

	// init prometheus
	go func() {
		var (
			err error
			mux *http.ServeMux
		)

		// creating new server instance so testing doesn't break multi-register of handlers
		mux = http.NewServeMux()
		sd.httpServer = new(http.Server)
		sd.httpServer.Addr = *addr
		sd.httpServer.Handler = mux

		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, "<h1>Netbox_SD Prometheus Metrics<h1><a href=\"/metrics\">see metrics here</a>\n")
		})

		mux.Handle("/metrics", promhttp.Handler())

		log.Printf("starting metrics http endpont on %s", sd.httpServer.Addr)

		err = sd.httpServer.ListenAndServe()

		if err != nil {
			log.Printf("failed to start metrics server: %v", err)
		}
	}()
}
