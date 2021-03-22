/*
© Copyright IBM Corporation 2018

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package metrics contains code to provide metrics for the queue manager
package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ot4i/ace-docker/common/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultPort = "9483"
)

var (
	metricsEnabled = false
	metricsServer  = &http.Server{Addr: ":" + defaultPort}
)

// GatherMetrics gathers metrics for the integration server
func GatherMetrics(serverName string, log logger.LoggerInterface) {
	log.Println("Gathering Metrics...")
	metricsEnabled = true

	err := startMetricsGathering(serverName, log)
	if err != nil {
		log.Errorf("Metrics Error: %s", err.Error())
		StopMetricsGathering()
	}
}

// startMetricsGathering starts gathering metrics for the integration server
func startMetricsGathering(serverName string, log logger.LoggerInterface) error {

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Metrics Error: %v", r)
		}
	}()

	log.Println("Starting metrics gathering")

	// Start processing metrics
	go processMetrics(log, serverName)

	// Wait for metrics to be ready before starting the Prometheus handler
	<-startChannel

	// Register metrics
	metricsExporter := newExporter(serverName, log)
	err := prometheus.Register(metricsExporter)
	if err != nil {
		return fmt.Errorf("Failed to register metrics: %v", err)
	}

	// Setup HTTP server to handle requests from Prometheus
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Status: METRICS ACTIVE"))
	})

	go func() {
		err = metricsServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("Metrics Error: Failed to handle metrics request: %v", err)

			StopMetricsGathering()
		}
	}()

	return nil
}

// StopMetricsGathering stops gathering metrics for the integration server
func StopMetricsGathering() {

	if metricsEnabled {

		// Stop processing metrics
		stopChannel <- true

		// Shutdown HTTP server
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		metricsServer.Shutdown(timeout)
	}
}
