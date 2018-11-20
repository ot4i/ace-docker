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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

var normaliseTests = []struct {
	md       metricData
	val      int
	expected float64
}{
	{metricData{metricUnits: Microseconds}, 1, 0.000001},
	{metricData{metricUnits: Microseconds}, 2000, 0.002},
	{metricData{metricUnits: Microseconds}, 0, 0.0},
	{metricData{metricUnits: Microseconds}, -1, 0.0},
	{metricData{metricUnits: Seconds}, 1, 1.0},
	{metricData{metricUnits: Seconds}, 100, 100.0},
	{metricData{metricUnits: Seconds}, 0, 0.0},
	{metricData{metricUnits: Seconds}, -1, 0.0},
	{metricData{metricUnits: Bytes}, 1, 1.0},
	{metricData{metricUnits: Bytes}, 1999, 1999.0},
	{metricData{metricUnits: Bytes}, 0, 0.0},
	{metricData{metricUnits: Bytes}, -1, 0.0},
	{metricData{metricUnits: MegaBytes}, 1, 1048576.0},
	{metricData{metricUnits: MegaBytes}, 50, 52428800.0},
	{metricData{metricUnits: MegaBytes}, 0, 0.0},
	{metricData{metricUnits: MegaBytes}, -1, 0.0},
	{metricData{metricUnits: Count}, 1, 1.0},
	{metricData{metricUnits: Count}, 542, 542.0},
	{metricData{metricUnits: Count}, 0, 0.0},
	{metricData{metricUnits: Count}, -1, 0.0},
}

func TestNormalise(t *testing.T) {
	for _, tt := range normaliseTests {
		t.Run(fmt.Sprintf("%d:%d", tt.md.metricUnits, tt.val), func(t *testing.T) {
			actual := tt.md.Normalise(tt.val)
			if actual != tt.expected {
				t.Errorf("Expected %d of type %d to normalise to %f, but got %f", tt.val, tt.md.metricUnits, tt.expected, actual)
			}
		})
	}
}

func TestInitialiseMetrics(t *testing.T) {

	/*
	   Get all the metric maps
	*/
	msgFlowMetricNamesMap, msgFlowNodeMetricNamesMap := generateMetricNamesMap()
	jvmMetricNamesMap := generateResourceMetricNamesMap()

	/*
	   Merge all the metric name maps into a single map to iterate over
	*/
	metricMapArray := [3]map[string]metricLookup{msgFlowMetricNamesMap, msgFlowNodeMetricNamesMap, jvmMetricNamesMap}

	aggregatedMetricsMap := make(map[string]metricLookup)
	disabledMetricsMap := make(map[string]metricLookup)

	for _, m := range metricMapArray {
		for k, v := range m {
			if v.enabled {
				aggregatedMetricsMap[k] = v
			} else {
				disabledMetricsMap[k] = v
			}
		}
	}

	/*
	  Initialise the metrics and check that:
	  - All entries from the metric name maps that are enabled are included
	  - There are no entries that are not in the metric name maps
	*/
	metrics := initialiseMetrics(getTestLogger())

	t.Logf("Iterating over aggregated metric name maps")
	for k, v := range aggregatedMetricsMap {
		t.Logf("- %s", k)

		metric, ok := metrics.internal[k]
		if !ok {
			t.Error("Expected metric not found in map")
		} else {
			if metric.name != v.name {
				t.Errorf("Expected name=%s; actual %s", v.name, metric.name)
			}
			if metric.description != v.description {
				t.Errorf("Expected description=%s; actual %s", v.description, metric.description)
			}
			if metric.metricType != v.metricType {
				t.Errorf("Expected metricType=%v; actual %v", v.metricType, metric.metricType)
			}
			if metric.metricLevel != v.metricLevel {
				t.Errorf("Expected metricLevel=%v; actual %v", v.metricLevel, metric.metricLevel)
			}
			if metric.metricUnits != v.metricUnits {
				t.Errorf("Expected metricUnits=%v; actual %v", v.metricUnits, metric.metricUnits)
			}
			if len(metric.values) != 0 {
				t.Errorf("Expected values-size=%d; actual %d", 0, len(metric.values))
			}
		}
	}

	if len(metrics.internal) != len(aggregatedMetricsMap) {
		t.Errorf("Map contains unexpected metrics, map size=%d", len(metrics.internal))
	}

	t.Logf("Iterating over map of disabled metric names")
	for k, _ := range disabledMetricsMap {
		t.Logf("- %s", k)

		metric, ok := metrics.internal[k]
		if ok {
			t.Errorf("Unexpected metric (%s) found in map: %+v", metric.name, metric)
		}
	}
}

func TestParseMetrics_AccountingAndStatistics(t *testing.T) {
	log := getTestLogger()

	expectedValues := map[string]int{
		"MsgFlow/TotalElapsedTime":                                          310845,
		"MsgFlow/MaximumElapsedTime":                                        54772,
		"MsgFlow/MinimumElapsedTime":                                        49184,
		"MsgFlow/TotalCpuTime":                                              262984,
		"MsgFlow/MaximumCpuTime":                                            53386,
		"MsgFlow/MinimumCpuTime":                                            40644,
		"MsgFlow/TotalSizeOfInputMessages":                                  2376,
		"MsgFlow/MaximumSizeOfInputMessages":                                396,
		"MsgFlow/MinimumSizeOfInputMessages":                                396,
		"MsgFlow/TotalInputMessages":                                        6,
		"MsgFlow/TotalCPUTimeWaiting":                                       125000,
		"MsgFlow/TotalElapsedTimeWaiting":                                   18932397,
		"MsgFlow/NumberOfThreadsInPool":                                     1,
		"MsgFlow/TimesMaximumNumberOfThreadsReached":                        6,
		"MsgFlow/TotalNumberOfMQErrors":                                     0,
		"MsgFlow/TotalNumberOfMessagesWithErrors":                           0,
		"MsgFlow/TotalNumberOfErrorsProcessingMessages":                     0,
		"MsgFlow/TotalNumberOfTimeOutsWaitingForRepliesToAggregateMessages": 0,
		"MsgFlow/TotalNumberOfCommits":                                      0,
		"MsgFlow/TotalNumberOfBackouts":                                     0,
	}

	expectedNodeValues := map[string]int{
		"HTTP Input/MsgFlowNode/TotalElapsedTime":   62219,
		"HTTP Input/MsgFlowNode/MaximumElapsedTime": 11826,
		"HTTP Input/MsgFlowNode/MinimumElapsedTime": 8797,
		"HTTP Input/MsgFlowNode/TotalCpuTime":       20747,
		"HTTP Input/MsgFlowNode/MaximumCpuTime":     10440,
		"HTTP Input/MsgFlowNode/MinimumCpuTime":     1,
		"HTTP Input/MsgFlowNode/TotalInvocations":   6,
		"HTTP Input/MsgFlowNode/InputTerminals":     0,
		"HTTP Input/MsgFlowNode/OutputTerminals":    4,
		"HTTP Reply/MsgFlowNode/TotalElapsedTime":   248626,
		"HTTP Reply/MsgFlowNode/MaximumElapsedTime": 42946,
		"HTTP Reply/MsgFlowNode/MinimumElapsedTime": 37639,
		"HTTP Reply/MsgFlowNode/TotalCpuTime":       242237,
		"HTTP Reply/MsgFlowNode/MaximumCpuTime":     42946,
		"HTTP Reply/MsgFlowNode/MinimumCpuTime":     31250,
		"HTTP Reply/MsgFlowNode/TotalInvocations":   6,
		"HTTP Reply/MsgFlowNode/InputTerminals":     1,
		"HTTP Reply/MsgFlowNode/OutputTerminals":    2,
	}

	/*
	   Parse a JSON string into a StatisticsDataStruct
	*/
	accountingAndStatisticsString :=
		"{\"data\":{\"WMQIStatisticsAccounting\":{\"RecordType\":\"SnapShot\",\"RecordCode\":\"SnapShot\",\"MessageFlow\":{\"BrokerLabel\":\"integration_server\",\"BrokerUUID\":\"\",\"ExecutionGroupName\":\"testintegrationserver\",\"ExecutionGroupUUID\":\"00000000-0000-0000-0000-000000000000\",\"MessageFlowName\":\"msgflow1\",\"ApplicationName\":\"application1\",\"StartDate\":\"2018-08-30\",\"StartTime\":\"13:51:11.277\",\"GMTStartTime\":\"2018-08-30T12:51:11.277+00:00\",\"EndDate\":\"2018-08-30\",\"EndTime\":\"13:51:31.514\",\"GMTEndTime\":\"2018-08-30T12:51:31.514+00:00\",\"TotalElapsedTime\":310845,\"MaximumElapsedTime\":54772,\"MinimumElapsedTime\":49184,\"TotalCPUTime\":262984,\"MaximumCPUTime\":53386,\"MinimumCPUTime\":40644,\"CPUTimeWaitingForInputMessage\":125000,\"ElapsedTimeWaitingForInputMessage\":18932397,\"TotalInputMessages\":6,\"TotalSizeOfInputMessages\":2376,\"MaximumSizeOfInputMessages\":396,\"MinimumSizeOfInputMessages\":396,\"NumberOfThreadsInPool\":1,\"TimesMaximumNumberOfThreadsReached\":6,\"TotalNumberOfMQErrors\":0,\"TotalNumberOfMessagesWithErrors\":0,\"TotalNumberOfErrorsProcessingMessages\":0,\"TotalNumberOfTimeOutsWaitingForRepliesToAggregateMessages\":0,\"TotalNumberOfCommits\":0,\"TotalNumberOfBackouts\":0,\"AccountingOrigin\":\"Anonymous\"},\"NumberOfThreads\":0,\"ThreadStatistics\":[],\"NumberOfNodes\":2,\"Nodes\":[{\"Label\":\"HTTP Input\",\"Type\":\"WSInputNode\",\"TotalElapsedTime\":62219,\"MaximumElapsedTime\":11826,\"MinimumElapsedTime\":8797,\"TotalCPUTime\":20747,\"MaximumCPUTime\":10440,\"MinimumCPUTime\":1,\"CountOfInvocations\":6,\"NumberOfInputTerminals\":0,\"NumberOfOutputTerminals\":4,\"TerminalStatistics\":[]},{\"Label\":\"HTTP Reply\",\"Type\":\"WSReplyNode\",\"TotalElapsedTime\":248626,\"MaximumElapsedTime\":42946,\"MinimumElapsedTime\":37639,\"TotalCPUTime\":242237,\"MaximumCPUTime\":42946,\"MinimumCPUTime\":31250,\"CountOfInvocations\":6,\"NumberOfInputTerminals\":1,\"NumberOfOutputTerminals\":2,\"TerminalStatistics\":[]}]}},\"event\":2}"

	var sds StatisticsDataStruct

	unmarshallError := json.Unmarshal([]byte(accountingAndStatisticsString), &sds)

	if unmarshallError != nil {
		t.Errorf("Error parsing json: %e", unmarshallError)
		return
	}

	mm, parseError := parseMetrics(log, &sds)
	if parseError != nil {
		t.Errorf("Error parsing metrics: %e", parseError)
		return
	}
	msgFlowMetricNamesMap, msgFlowNodeMetricNamesMap := generateMetricNamesMap()

	for k, v := range msgFlowMetricNamesMap {
		if v.enabled == false {
			metric, ok := mm.internal[k]
			if ok {
				t.Errorf("Unexpected metric (%s) found in map: %+v", k, metric)
			}
		} else {
			metric, ok := mm.internal[k]
			if !ok {
				t.Errorf("Missing expected metric (%s)", k)
			}
			if metric.name != v.name {
				t.Errorf("Expected name=%s; actual %s", v.name, metric.name)
			}
			if metric.description != v.description {
				t.Errorf("Expected description=%s; actual %s", v.description, metric.description)
			}
			if metric.metricType != v.metricType {
				t.Errorf("Expected metricType=%v; actual %v", v.metricType, metric.metricType)
			}
			if metric.metricLevel != v.metricLevel {
				t.Errorf("Expected metricLevel=%v; actual %v", v.metricLevel, metric.metricLevel)
			}
			if metric.metricUnits != v.metricUnits {
				t.Errorf("Expected metricUnits=%v; actual %v", v.metricUnits, metric.metricUnits)
			}
			if len(metric.values) != 1 {
				t.Errorf("Expected values-size=%d; actual %d", 1, len(metric.values))
			}

			for vk, values := range metric.values {
				if vk != "Anonymous_application1_msgflow1" {
					t.Errorf("Expected values key=%s; actual %s", "Anonymous_application1_msgflow1", vk)
				}

				for label, labelValue := range values.labels {
					switch label {
					case serverLabel:
						if labelValue != "testintegrationserver" {
							t.Errorf("Expected server label=%s; actual %s", "testintegrationserver", labelValue)
						}
					case applicationLabel:
						if labelValue != "application1" {
							t.Errorf("Expected application label=%s; actual %s", "application1", labelValue)
						}
					case msgflowPrefix:
						if labelValue != "msgflow1" {
							t.Errorf("Expected msgflow label=%s; actual %s", "msgflow1", labelValue)
						}
					case originLabel:
						if labelValue != "Anonymous" {
							t.Errorf("Expected origin label=%s; actual %s", "Anonymous", labelValue)
						}
					default:
						t.Errorf("Unexpected label (%s) found for metric: %s", label, metric.name)
					}
				}

				if values.value != metric.Normalise(expectedValues[k]) {
					t.Errorf("Expected %s value=%f; actual %f", metric.name, metric.Normalise(expectedValues[k]), values.value)
				}
			}
		}
	}

	for k, v := range msgFlowNodeMetricNamesMap {
		if v.enabled == false {
			metric, ok := mm.internal[k]
			if ok {
				t.Errorf("Unexpected metric (%s) found in map: %+v", metric.name, metric)
			}
		} else {
			metric, ok := mm.internal[k]
			if !ok {
				t.Errorf("Missing expected metric (%s)", k)
			}
			if metric.name != v.name {
				t.Errorf("Expected name=%s; actual %s", v.name, metric.name)
			}
			if metric.description != v.description {
				t.Errorf("Expected description=%s; actual %s", v.description, metric.description)
			}
			if metric.metricType != v.metricType {
				t.Errorf("Expected metricType=%v; actual %v", v.metricType, metric.metricType)
			}
			if metric.metricLevel != v.metricLevel {
				t.Errorf("Expected metricLevel=%v; actual %v", v.metricLevel, metric.metricLevel)
			}
			if metric.metricUnits != v.metricUnits {
				t.Errorf("Expected metricUnits=%v; actual %v", v.metricUnits, metric.metricUnits)
			}
			if len(metric.values) != 2 {
				t.Errorf("Expected values-size=%d; actual %d", 2, len(metric.values))
			}

			for vk, values := range metric.values {

				nodeName, ok := values.labels["msgflownode"]
				if !ok {
					t.Errorf("Missing label for msgflownode name: %s", k)
				}

				if vk != "Anonymous_application1_msgflow1_"+nodeName {
					t.Errorf("Expected values key=%s; actual %s", "Anonymous_application1_msgflow1_"+nodeName, vk)
				}

				for label, labelValue := range values.labels {
					switch label {
					case serverLabel:
						if labelValue != "testintegrationserver" {
							t.Errorf("Expected server label=%s; actual %s", "testintegrationserver", labelValue)
						}
					case applicationLabel:
						if labelValue != "application1" {
							t.Errorf("Expected application label=%s; actual %s", "application1", labelValue)
						}
					case msgflowPrefix:
						if labelValue != "msgflow1" {
							t.Errorf("Expected msgflow label=%s; actual %s", "msgflow1", labelValue)
						}
					case originLabel:
						if labelValue != "Anonymous" {
							t.Errorf("Expected origin label=%s; actual %s", "Anonymous", labelValue)
						}
					case msgflownodeLabel:
						// TODO: Check label value
					case msgflownodeTypeLabel:
						// TODO: Check label value
					default:
						t.Errorf("Unexpected label (%s) found for metric: %s", label, metric.name)
					}
				}

				if values.value != metric.Normalise(expectedNodeValues[nodeName+"/"+k]) {
					t.Errorf("Expected %s value=%f; actual %f", metric.name, metric.Normalise(expectedValues[nodeName+"/"+k]), values.value)
				}
			}
		}
	}
}

func TestParseMetrics_ResourceStatistics(t *testing.T) {
	log := getTestLogger()

	expectedValues := map[string]int{
		"JVM/Summary/InitialMemoryInMB":                   305,
		"JVM/Summary/UsedMemoryInMB":                      40,
		"JVM/Summary/CommittedMemoryInMB":                 314,
		"JVM/Summary/MaxMemoryInMB":                       -1,
		"JVM/Summary/CumulativeGCTimeInSeconds":           0,
		"JVM/Summary/CumulativeNumberOfGCCollections":     132,
		"JVM/Heap/InitialMemoryInMB":                      32,
		"JVM/Heap/UsedMemoryInMB":                         18,
		"JVM/Heap/CommittedMemoryInMB":                    34,
		"JVM/Heap/MaxMemoryInMB":                          256,
		"JVM/Native/InitialMemoryInMB":                    273,
		"JVM/Native/UsedMemoryInMB":                       22,
		"JVM/Native/CommittedMemoryInMB":                  280,
		"JVM/Native/MaxMemoryInMB":                        -1,
		"JVM/ScavengerGC/CumulativeGCTimeInSeconds":       0,
		"JVM/ScavengerGC/CumulativeNumberOfGCCollections": 131,
		"JVM/GlobalGC/CumulativeGCTimeInSeconds":          0,
		"JVM/GlobalGC/CumulativeNumberOfGCCollections":    1,
	}

	resourceStatisticsString := "{\"data\":{\"ResourceStatistics\":{\"brokerLabel\":\"integration_server\",\"brokerUUID\":\"\",\"executionGroupName\":\"testintegrationserver\",\"executionGroupUUID\":\"00000000-0000-0000-0000-000000000000\",\"collectionStartDate\":\"2018-08-28\",\"collectionStartTime\":\"21:13:33\",\"startDate\":\"2018-08-30\",\"startTime\":\"13:50:54\",\"endDate\":\"2018-08-30\",\"endTime\":\"13:51:15\",\"timezone\":\"Europe/London\",\"ResourceType\":[{\"name\":\"JVM\",\"resourceIdentifier\":[{\"name\":\"summary\",\"InitialMemoryInMB\":305,\"UsedMemoryInMB\":40,\"CommittedMemoryInMB\":314,\"MaxMemoryInMB\":-1,\"CumulativeGCTimeInSeconds\":0,\"CumulativeNumberOfGCCollections\":132},{\"name\":\"Heap Memory\",\"InitialMemoryInMB\":32,\"UsedMemoryInMB\":18,\"CommittedMemoryInMB\":34,\"MaxMemoryInMB\":256},{\"name\":\"Non-Heap Memory\",\"InitialMemoryInMB\":273,\"UsedMemoryInMB\":22,\"CommittedMemoryInMB\":280,\"MaxMemoryInMB\":-1},{\"name\":\"Garbage Collection - scavenge\",\"CumulativeGCTimeInSeconds\":0,\"CumulativeNumberOfGCCollections\":131},{\"name\":\"Garbage Collection - global\",\"CumulativeGCTimeInSeconds\":0,\"CumulativeNumberOfGCCollections\":1}]}]}},\"event\":0}"

	var sds StatisticsDataStruct

	unmarshallError := json.Unmarshal([]byte(resourceStatisticsString), &sds)

	if unmarshallError != nil {
		t.Errorf("Error parsing json: %e", unmarshallError)
	}

	mm, parseError := parseMetrics(log, &sds)
	if parseError != nil {
		t.Errorf("Error parsing metrics: %e", parseError)
		return
	}
	jvmMetricNamesMap := generateResourceMetricNamesMap()

	for k, v := range jvmMetricNamesMap {
		if v.enabled == false {
			metric, ok := mm.internal[k]
			if ok {
				t.Errorf("Unexpected metric (%s) found in map: %+v", metric.name, metric)
			}
		} else {
			metric, ok := mm.internal[k]
			if !ok {
				t.Errorf("Missing expected metric (%s)", k)
			}
			if metric.name != v.name {
				t.Errorf("Expected name=%s; actual %s", v.name, metric.name)
			}
			if metric.description != v.description {
				t.Errorf("Expected description=%s; actual %s", v.description, metric.description)
			}
			if metric.metricType != v.metricType {
				t.Errorf("Expected metricType=%v; actual %v", v.metricType, metric.metricType)
			}
			if metric.metricLevel != v.metricLevel {
				t.Errorf("Expected metricLevel=%v; actual %v", v.metricLevel, metric.metricLevel)
			}
			if metric.metricUnits != v.metricUnits {
				t.Errorf("Expected metricUnits=%v; actual %v", v.metricUnits, metric.metricUnits)
			}
			if len(metric.values) != 1 {
				t.Errorf("Expected values-size=%d; actual %d", 1, len(metric.values))
			}

			for _, values := range metric.values {
				for label, labelValue := range values.labels {
					switch label {
					case serverLabel:
						if labelValue != "testintegrationserver" {
							t.Errorf("Expected server label=%s; actual %s", "testintegrationserver", labelValue)
						}
					default:
						t.Errorf("Unexpected label (%s) found for metric: %s", label, metric.name)
					}
				}

				if values.value != metric.Normalise(expectedValues[k]) {
					t.Errorf("Expected %s value=%f; actual %f", metric.name, metric.Normalise(expectedValues[k]), values.value)
				}
			}
		}
	}
}

var updateTests = []struct {
	mType    MetricType
	mValue1  float64
	mValue2  float64
	expected float64
}{
	{Total, 9.0, 5.0, 14.0},
	{Total, 0.0, 5.0, 5.0},
	{Total, 32.0, 0.0, 32.0},
	{Total, 0.0, 0.0, 0.0},
	{Minimum, 0.0, 0.0, 0.0},
	{Minimum, 1.0, 0.0, 0.0},
	{Minimum, 0.0, 3.4, 0.0},
	{Minimum, 5.0, 23.7, 5.0},
	{Minimum, 31.0, 30.9, 30.9},
	{Maximum, 0.0, 0.0, 0.0},
	{Maximum, 1.0, 0.0, 1.0},
	{Maximum, 0.0, 3.4, 3.4},
	{Maximum, 5.0, 23.7, 23.7},
	{Maximum, 31.0, 30.9, 31.0},
	{Current, 0.0, 0.0, 0.0},
	{Current, 1.0, 0.0, 0.0},
	{Current, 0.0, 3.4, 3.4},
	{Current, 5.0, 23.7, 23.7},
	{Current, 31.0, 30.9, 30.9},
}

func TestUpdateMetrics_Simple(t *testing.T) {
	log := getTestLogger()

	for _, tt := range updateTests {
		mm1 := NewMetricsMap()
		mm2 := NewMetricsMap()

		m1 := metricData{metricType: tt.mType}
		m1.values = make(map[string]*Metric)
		m1.values["test_val_key"] = &Metric{labels: prometheus.Labels{serverLabel: "test_server"}, value: tt.mValue1}
		mm1.internal["test_mm_key"] = &m1

		m2 := metricData{metricType: tt.mType}
		m2.values = make(map[string]*Metric)
		m2.values["test_val_key"] = &Metric{labels: prometheus.Labels{serverLabel: "test_server"}, value: tt.mValue2}
		mm2.internal["test_mm_key"] = &m2

		t.Run(fmt.Sprintf("%d", tt.mType), func(t *testing.T) {
			updateMetrics(log, mm1, mm2)
			actual := mm1.internal["test_mm_key"].values["test_val_key"].value
			if actual != tt.expected {
				t.Errorf("Expected update of type:%d with values %f and %f to result in a new value of %f, but actual=%f", tt.mType, tt.mValue1, tt.mValue2, tt.expected, actual)
			}
		})
	}
}

func TestUpdateMetrics_NewValue(t *testing.T) {
	log := getTestLogger()

	mm1 := NewMetricsMap()
	mm2 := NewMetricsMap()

	m1 := metricData{metricType: Total}
	m1.values = make(map[string]*Metric)
	m1.values["non_existent_test_key"] = &Metric{labels: prometheus.Labels{serverLabel: "test_server"}, value: 4.0}
	mm1.internal["test_mm_key"] = &m1

	m2 := metricData{metricType: Total}
	m2.values = make(map[string]*Metric)
	m2.values["existing_test_key"] = &Metric{labels: prometheus.Labels{serverLabel: "test_server"}, value: 7.1}
	mm2.internal["test_mm_key"] = &m2

	updateMetrics(log, mm1, mm2)
	actual1 := mm1.internal["test_mm_key"].values["existing_test_key"].value
	if actual1 != 7.1 {
		t.Errorf("Value for metric key existing_test_key expected:%f, actual=%f", 7.1, actual1)
	}

	actual2 := mm1.internal["test_mm_key"].values["non_existent_test_key"].value
	if actual2 != 4.0 {
		t.Errorf("Value for metric key non_existent_test_key expected:%f, actual=%f", 4.0, actual2)
	}
}

func TestUpdateMetrics_NewKey(t *testing.T) {
	log := getTestLogger()

	mm1 := NewMetricsMap()
	mm2 := NewMetricsMap()

	m1 := metricData{metricType: Total}
	m1.values = make(map[string]*Metric)
	m1.values["test_key"] = &Metric{labels: prometheus.Labels{serverLabel: "test_server"}, value: 4.0}
	mm1.internal["new_key"] = &m1

	m2 := metricData{metricType: Total}
	m2.values = make(map[string]*Metric)
	m2.values["test_key"] = &Metric{labels: prometheus.Labels{serverLabel: "test_server"}, value: 7.1}
	mm2.internal["old_key"] = &m2

	updateMetrics(log, mm1, mm2)
	actual1 := mm1.internal["old_key"].values["test_key"].value
	if actual1 != 7.1 {
		t.Errorf("Value for metric key existing_test_key expected:%f, actual=%f", 7.1, actual1)
	}

	actual2 := mm1.internal["new_key"].values["test_key"].value
	if actual2 != 4.0 {
		t.Errorf("Value for metric key non_existent_test_key expected:%f, actual=%f", 4.0, actual2)
	}
}
