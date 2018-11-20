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
package metrics

import (
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/ot4i/ace-docker/internal/logger"
)

func getTestLogger() *logger.Logger {
	log, _ := logger.NewLogger(os.Stdout, false, false, "test")
	return log
}

func TestDescribe(t *testing.T) {
	log := getTestLogger()

	ch := make(chan *prometheus.Desc)
	go func() {
		exporter := newExporter("serverName", log)
		exporter.Describe(ch)
	}()

	collect := <-requestChannel
	if collect {
		t.Errorf("Received unexpected collect request")
	}

	metrics := NewMetricsMap()

	countMetric := metricData{
		name:        "test_count_metric",
		description: "This is a test counter metric",
		metricType:  Total,
		metricUnits: Count,
		metricLevel: MsgFlowLevel,
	}
	countMetric.values = make(map[string]*Metric)

	gaugeMetric := metricData{
		name:        "test_gauge_metric",
		description: "This is a test gauge metric",
		metricType:  Current,
		metricUnits: Count,
		metricLevel: Resource,
	}
	gaugeMetric.values = make(map[string]*Metric)

	metrics.internal["count_metrics"] = &countMetric
	metrics.internal["gauge_metrics"] = &gaugeMetric

	responseChannel <- metrics

	expectedDesc1 := "Desc{fqName: \"ibmace_test_count_metric\", help: \"This is a test counter metric\", constLabels: {}, variableLabels: [msgflow application server accountingorigin]}"
	expectedDesc2 := "Desc{fqName: \"ibmace_test_gauge_metric\", help: \"This is a test gauge metric\", constLabels: {}, variableLabels: [server]}"

	var found1, found2 bool = false, false

	for i := 0; i < len(metrics.internal); i++ {
		prometheusDesc := <-ch
		actualDesc := prometheusDesc.String()

		if actualDesc != expectedDesc1 && actualDesc != expectedDesc2 {
			t.Errorf("Expected a value of either\n- %s OR\n- %s\n\nActual value was - %s", expectedDesc1, expectedDesc2, actualDesc)
			return
		}

		if actualDesc == expectedDesc1 {
			if found1 {
				t.Errorf("Duplicate metrics sent over channel, was only expected once\n- %s\n", expectedDesc1)
			} else {
				found1 = true
			}
		}

		if actualDesc == expectedDesc2 {
			if found2 {
				t.Errorf("Duplicate metrics sent over channel, was only expected once\n- %s\n", expectedDesc2)
			} else {
				found2 = true
			}
		}
	}

	if !found1 {
		t.Errorf("Expected to find value\n- %s", expectedDesc1)
	}

	if !found2 {
		t.Errorf("Expected to find value\n- %s", expectedDesc2)
	}
}

func TestCollect_Counters(t *testing.T) {
	log := getTestLogger()

	exporter := newExporter("serverName", log)

	exporter.counterMap["count_metrics"] = createCounterVec("test_count_metric", "This is a test counter metric", Resource)

	metrics := NewMetricsMap()

	countMetric := metricData{
		name:        "test_count_metric",
		description: "This is a test counter metric",
		metricType:  Total,
		metricUnits: Count,
		metricLevel: Resource,
	}
	countMetric.values = make(map[string]*Metric)
	metrics.internal["count_metrics"] = &countMetric

	countValues := []float64{0.0, 4.0, 5.0, 0.0, 3.0}
	expected := 0.0

	// Call collect several times and ensure the values are reset and counter is incremented as expected
	for _, countValue := range countValues {
		expected += countValue

		ch := make(chan prometheus.Metric)
		go func() {
			exporter.Collect(ch)
			close(ch)
		}()

		collect := <-requestChannel
		if !collect {
			t.Errorf("Received unexpected describe request")
		}

		countMetric.values["Test1"] = &Metric{labels: prometheus.Labels{serverLabel: "test server1"}, value: countValue}

		responseChannel <- metrics
		<-ch
		prometheusMetric := dto.Metric{}
		exporter.counterMap["count_metrics"].WithLabelValues("test server1").Write(&prometheusMetric)
		actual := prometheusMetric.GetCounter().GetValue()

		if actual != expected {
			t.Errorf("Expected value=%f; actual=%f", expected, actual)
		}

		if len(countMetric.values) != 0 {
			t.Errorf("Counter values should be reset after collect: %+v", countMetric.values)
		}
	}
}

func TestCollect_Gauges(t *testing.T) {
	log := getTestLogger()

	exporter := newExporter("TestServer", log)

	exporter.gaugeMap["gauge_metrics"] = createGaugeVec("test_gauge_metric", "This is a test gauge metric", Resource)

	metrics := NewMetricsMap()

	gaugeMetric := metricData{
		name:        "test_gauge_metric",
		description: "This is a test gauge metric",
		metricType:  Current,
		metricUnits: Count,
		metricLevel: Resource,
	}
	gaugeMetric.values = make(map[string]*Metric)
	metrics.internal["gauge_metrics"] = &gaugeMetric

	gaugeValues := []float64{0.0, 4.0, 5.0, 0.0, 3.0}

	// Call collect several times and ensure the values are reset and counter is incremented as expected
	for _, gaugeValue := range gaugeValues {
		expected := gaugeValue

		ch := make(chan prometheus.Metric)
		go func() {
			exporter.Collect(ch)
			close(ch)
		}()

		collect := <-requestChannel
		if !collect {
			t.Errorf("Received unexpected describe request")
		}

		gaugeMetric.values["Test"] = &Metric{labels: prometheus.Labels{serverLabel: "TestServer"}, value: gaugeValue}

		responseChannel <- metrics
		<-ch
		prometheusMetric := dto.Metric{}
		exporter.gaugeMap["gauge_metrics"].WithLabelValues("TestServer").Write(&prometheusMetric)
		actual := prometheusMetric.GetGauge().GetValue()

		if actual != expected {
			t.Errorf("Expected value=%f; actual=%f", expected, actual)
		}

		if len(gaugeMetric.values) != 1 {
			t.Errorf("Gauge values should not be reset after collect: %+v", gaugeMetric.values)
		}
	}
}

func TestCreateCounterVec_msgFlow(t *testing.T) {

	ch := make(chan *prometheus.Desc)
	counterVec := createCounterVec("MetricName", "MetricDescription", MsgFlowLevel)
	go func() {
		counterVec.Describe(ch)
	}()
	description := <-ch

	expected := "Desc{fqName: \"ibmace_MetricName\", help: \"MetricDescription\", constLabels: {}, variableLabels: [msgflow application server accountingorigin]}"
	actual := description.String()
	if actual != expected {
		t.Errorf("Expected value=%s; actual %s", expected, actual)
	}
}

func TestCreateCounterVec_MsgFlowNode(t *testing.T) {

	ch := make(chan *prometheus.Desc)
	counterVec := createCounterVec("MetricName", "MetricDescription", MsgFlowNodeLevel)
	go func() {
		counterVec.Describe(ch)
	}()
	description := <-ch

	expected := "Desc{fqName: \"ibmace_MetricName\", help: \"MetricDescription\", constLabels: {}, variableLabels: [msgflownode msgflownodetype msgflow application server accountingorigin]}"
	actual := description.String()
	if actual != expected {
		t.Errorf("Expected value=%s; actual %s", expected, actual)
	}
}

func TestCreateCounterVec_Resource(t *testing.T) {

	ch := make(chan *prometheus.Desc)
	counterVec := createCounterVec("MetricName", "MetricDescription", Resource)
	go func() {
		counterVec.Describe(ch)
	}()
	description := <-ch

	expected := "Desc{fqName: \"ibmace_MetricName\", help: \"MetricDescription\", constLabels: {}, variableLabels: [server]}"
	actual := description.String()
	if actual != expected {
		t.Errorf("Expected value=%s; actual %s", expected, actual)
	}
}

func TestCreateGaugeVec_MsgFlow(t *testing.T) {

	ch := make(chan *prometheus.Desc)
	gaugeVec := createGaugeVec("MetricName", "MetricDescription", MsgFlowLevel)
	go func() {
		gaugeVec.Describe(ch)
	}()
	description := <-ch

	expected := "Desc{fqName: \"ibmace_MetricName\", help: \"MetricDescription\", constLabels: {}, variableLabels: [msgflow application server accountingorigin]}"
	actual := description.String()
	if actual != expected {
		t.Errorf("Expected value=%s; actual %s", expected, actual)
	}
}

func TestCreateGaugeVec_MsgFlowNode(t *testing.T) {

	ch := make(chan *prometheus.Desc)
	gaugeVec := createGaugeVec("MetricName", "MetricDescription", MsgFlowNodeLevel)
	go func() {
		gaugeVec.Describe(ch)
	}()
	description := <-ch

	expected := "Desc{fqName: \"ibmace_MetricName\", help: \"MetricDescription\", constLabels: {}, variableLabels: [msgflownode msgflownodetype msgflow application server accountingorigin]}"
	actual := description.String()
	if actual != expected {
		t.Errorf("Expected value=%s; actual %s", expected, actual)
	}
}

func TestCreateGaugeVec_Resource(t *testing.T) {

	ch := make(chan *prometheus.Desc)
	gaugeVec := createGaugeVec("MetricName", "MetricDescription", Resource)
	go func() {
		gaugeVec.Describe(ch)
	}()
	description := <-ch

	expected := "Desc{fqName: \"ibmace_MetricName\", help: \"MetricDescription\", constLabels: {}, variableLabels: [server]}"
	actual := description.String()
	if actual != expected {
		t.Errorf("Expected value=%s; actual %s", expected, actual)
	}
}
