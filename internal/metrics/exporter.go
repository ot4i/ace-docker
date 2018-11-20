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
	"github.com/ot4i/ace-docker/internal/logger"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace            = "ibmace"
	msgflowPrefix        = "msgflow"
	msgflownodePrefix    = "msgflownode"
	serverLabel          = "server"
	applicationLabel     = "application"
	msgflowLabel         = "msgflow"
	msgflownodeLabel     = "msgflownode"
	msgflownodeTypeLabel = "msgflownodetype"
	originLabel          = "accountingorigin"
)

type exporter struct {
	serverName   string
	counterMap   map[string]*prometheus.CounterVec
	gaugeMap     map[string]*prometheus.GaugeVec
	firstCollect bool
	log          *logger.Logger
}

func newExporter(serverName string, log *logger.Logger) *exporter {
	return &exporter{
		serverName: serverName,
		counterMap: make(map[string]*prometheus.CounterVec),
		gaugeMap:   make(map[string]*prometheus.GaugeVec),
		log:        log,
	}
}

// Describe provides details of all available metrics
func (e *exporter) Describe(ch chan<- *prometheus.Desc) {

	requestChannel <- false
	response := <-responseChannel

	response.Lock()
	defer response.Unlock()

	for key, metric := range response.internal {

		if metric.metricType == Total {
			// For delta type metrics - allocate a Prometheus Counter
			counterVec := createCounterVec(metric.name, metric.description, metric.metricLevel)
			e.counterMap[key] = counterVec

			// Describe metric
			counterVec.Describe(ch)
		} else {
			// For non-delta type metrics - allocate a Prometheus Gauge
			gaugeVec := createGaugeVec(metric.name, metric.description, metric.metricLevel)
			e.gaugeMap[key] = gaugeVec

			// Describe metric
			gaugeVec.Describe(ch)
		}

	}
}

// Collect is called at regular intervals to provide the current metric data
func (e *exporter) Collect(ch chan<- prometheus.Metric) {

	requestChannel <- true
	response := <-responseChannel

	response.Lock()
	defer response.Unlock()

	for key, metric := range response.internal {
		if metric.metricType == Total {
			// For delta type metrics - update their Prometheus Counter
			counterVec := e.counterMap[key]

			// Populate Prometheus Counter with metric values
			for _, value := range metric.values {
				var err error
				var counter prometheus.Counter

				counter, err = counterVec.GetMetricWith(value.labels)

				if err == nil {
					counter.Add(value.value)
				} else {
					e.log.Errorf("Metrics Error: %s", err.Error())
				}
			}

			// Collect metric and reset cached values
			counterVec.Collect(ch)
			response.internal[key].values = make(map[string]*Metric)
		} else {
			// For non-delta type metrics - reset their Prometheus Gauge
			gaugeVec := e.gaugeMap[key]
			gaugeVec.Reset()

			for _, value := range metric.values {
				var err error
				var gauge prometheus.Gauge

				gauge, err = gaugeVec.GetMetricWith(value.labels)

				if err == nil {
					gauge.Set(value.value)
				} else {
					e.log.Errorf("Metrics Error: %s", err.Error())
				}
			}

			// Collect metric
			gaugeVec.Collect(ch)
		}
	}
}

// createCounterVec returns a Prometheus CounterVec populated with metric details
func createCounterVec(name, description string, metricLevel MetricLevel) *prometheus.CounterVec {

	labels := getVecDetails(metricLevel)

	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      name,
			Help:      description,
		},
		labels,
	)
	return counterVec
}

// createGaugeVec returns a Prometheus GaugeVec populated with metric details
func createGaugeVec(name, description string, metricLevel MetricLevel) *prometheus.GaugeVec {

	labels := getVecDetails(metricLevel)

	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      name,
			Help:      description,
		},
		labels,
	)
	return gaugeVec
}

// getVecDetails returns the required prefix and labels for a metric
func getVecDetails(metricLevel MetricLevel) (labels []string) {

	//TODO: What if messageflow is in a library?
	if metricLevel == MsgFlowLevel {
		labels = []string{msgflowLabel, applicationLabel, serverLabel, originLabel}
	} else if metricLevel == MsgFlowNodeLevel {
		labels = []string{msgflownodeLabel, msgflownodeTypeLabel, msgflowLabel, applicationLabel, serverLabel, originLabel}
	} else if metricLevel == Resource {
		labels = []string{serverLabel}
	}

	return labels
}
