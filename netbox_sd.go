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
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/4xoc/netbox_sd/netbox"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"gopkg.in/yaml.v3"
)

const (
	WorkerSleepTimeMS = 500
)

type netboxSD struct {
	cfg        *Config
	api        netbox.ClientIface
	httpServer *http.Server
}

var (
	// All cmd flags come here.
	cfgFile     = flag.String("config.file", "config.yml", "config file path")
	showVersion = flag.Bool("version", false, "show version information")
	debug       = flag.Bool("debug", false, "enable debug output")
	promListen  = flag.String("web.listen", "[::]:9099", "prometheus metrics listen address")

	// SD is the single global instance of netboxSD to manage all groups.
	sd *netboxSD = new(netboxSD)

	// subsituted on build time
	version string
	commit  string
	date    string
)

func init() {
	// Who doesn't like more details in their logs?
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Version: %s (compiled on %s with commit %s)\n", version, date, commit)

	flag.Usage = func() {
		fmt.Println("Usage: netbox_sd [parameters]\n\nParameters:")
		flag.PrintDefaults()
		fmt.Println("\n" + `MIT License - Copyright (c) 2024 WIIT AG`)
	}
}

func main() {
	var (
		err error
		i   int
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s (compiled on %s with commit %s)\n", version, date, commit)
		os.Exit(0)
	}

	sd.serveMetrics(promListen)

	log.Printf("loading config")

	sd.cfg, err = readConfigFile(*cfgFile)
	if err != nil {
		log.Printf("failed to load config file: %v", err)
		os.Exit(1)
	}

	sd.api, err = netbox.New(sd.cfg.BaseURL, sd.cfg.Token, PrometheusNameSpace, true, sd.cfg.AllowInsecure)
	if err != nil {
		log.Printf("failed to initialize new api client")
		os.Exit(1)
	}

	if *debug {
		sd.api.HTTPTracing(true)
	}

	err = sd.api.VerifyConnectivity()
	if err != nil {
		log.Printf("failed to verify connectivity to Netbox: %v", err)
		os.Exit(1)
	} else {
		log.Printf("connection to Netbox successful")
	}

	// At this point the config has been read and been through a basic validation. The Netbox API client is initialized
	// and the provided baseURL and token seem fine. Now we can start with the actual data gathering.

	promGroups.Set(float64(len(sd.cfg.Groups)))

	// Start an independent worker thread per group. This makes tracking the individual scanInterval much easier and who
	// doesn't like goroutines?
	for i = range sd.cfg.Groups {
		log.Printf("starting worker for group %s", sd.cfg.Groups[i].File)
		go sd.worker(sd.cfg.Groups[i])
	}

	// wait until the end of times
	select {}
}

// Worker performs all necessary steps to fetch targets based on the group's configuration markers and writes those
// targets into a file that can be picked up by Prometheus' file_sd.
func (sd *netboxSD) worker(group *Group) {
	var (
		// init last run with a time that is sure to trigger a scan on first iteration
		lastRun  time.Time = time.Now().Add(-group.ScanInterval)
		runStart time.Time
		failed   bool
		err      error
		targets  []*targetgroup.Group
		data     []byte
	)

	for {
		if time.Since(lastRun) >= group.ScanInterval {
			if *debug {
				log.Printf("new scan for group %s\n", group.File)
			}

			// reset vars
			runStart = time.Now()
			failed = false

			switch group.Type {
			case GroupTypeService:
				targets, err = sd.getTargetsByService(group)
				if err != nil {
					log.Printf("getting targets for group %s failed: %s", group.File, err.Error())
					failed = true
				}

			case GroupTypeDeviceTag:
				targets, err = sd.getTargetsByDeviceTag(group)
				if err != nil {
					log.Printf("getting targets for group %s failed: %s", group.File, err.Error())
					failed = true
				}

			case GroupTypeInterfaceTag:
				targets, err = sd.getTargetsByInterfaceTag(group)
				if err != nil {
					log.Printf("getting targets for group %s failed: %s", group.File, err.Error())
					failed = true
				}
			}

			if !failed {
				// NOTE: Unfortunately only YAML is a valid option here since there is no proper way to marshal JSON. See this
				// issue: https://github.com/prometheus/prometheus/pull/6691.
				data, err = yaml.Marshal(targets)
				if err != nil {
					// This should never happen unless there is as bug in Prometheus. This panicing here so this get's picked up.
					log.Panicf("parsing targets to yaml failed: %v", err)
				}

				err = os.WriteFile(group.File, data, 0664)
				if err != nil {
					log.Printf("failed to write file %s: %v", group.File, err)
					failed = true
				} else {
					// Update target count; otherwise we report the old value as nothing has changed.
					promTargetCount.
						With(prometheus.Labels{
							"group": group.File,
						}).
						Set(float64(len(targets)))
				}
			} else {
				promUpdateError.
					With(prometheus.Labels{
						"group": group.File,
					}).
					Inc()
			}

			// Update lastRun time to track next iteration.
			lastRun = time.Now()

			promUpdateDuration.
				With(prometheus.Labels{
					"group": group.File,
				}).
				Set(float64(time.Since(runStart).Nanoseconds()))

			promUpdateTime.
				With(prometheus.Labels{
					"group": group.File,
				}).Set(float64(time.Now().Unix()))
		}

		time.Sleep(WorkerSleepTimeMS * time.Millisecond)
	}
}
