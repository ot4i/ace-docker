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
	"testing"
)

func TestGenerateMetricNamesMap(t *testing.T) {

	msgFlowMetricNamesMap, msgFlowNodeMetricNamesMap := generateMetricNamesMap()

	if len(msgFlowMetricNamesMap) != 20 {
		t.Errorf("Expected mapping-size=%d; actual %d", 20, len(msgFlowMetricNamesMap))
	}

	if len(msgFlowNodeMetricNamesMap) != 9 {
		t.Errorf("Expected mapping-size=%d; actual %d", 9, len(msgFlowNodeMetricNamesMap))
	}

	for _, v := range msgFlowMetricNamesMap {
		if v.name == "" {
			t.Errorf("Name for metricLookup is empty string: %+v", v)
		}
		if v.description == "" {
			t.Errorf("Description for metricLookup is empty string: %+v", v)
		}
	}

	for _, v := range msgFlowNodeMetricNamesMap {
		if v.name == "" {
			t.Errorf("Name for metricLookup is empty string: %+v", v)
		}
		if v.description == "" {
			t.Errorf("Description for metricLookup is empty string: %+v", v)
		}
	}
}

func TestGenerateResourceMetricNamesMap(t *testing.T) {

	jvmMetricNamesMap := generateResourceMetricNamesMap()

	if len(jvmMetricNamesMap) != 18 {
		t.Errorf("Expected mapping-size=%d; actual %d", 18, len(jvmMetricNamesMap))
	}

	for _, v := range jvmMetricNamesMap {
		if v.name == "" {
			t.Errorf("Name for metricLookup is empty string: %+v", v)
		}
		if v.description == "" {
			t.Errorf("Description for metricLookup is empty string: %+v", v)
		}
	}
}

func TestNewJVMData(t *testing.T) {
	resourceStatisticsString := "{\"data\":{\"ResourceStatistics\":{\"brokerLabel\":\"integration_server\",\"brokerUUID\":\"\",\"executionGroupName\":\"websockettest\",\"executionGroupUUID\":\"00000000-0000-0000-0000-000000000000\",\"collectionStartDate\":\"2018-08-28\",\"collectionStartTime\":\"21:13:33\",\"startDate\":\"2018-08-30\",\"startTime\":\"13:50:54\",\"endDate\":\"2018-08-30\",\"endTime\":\"13:51:15\",\"timezone\":\"Europe/London\",\"ResourceType\":[{\"name\":\"JVM\",\"resourceIdentifier\":[{\"name\":\"summary\",\"InitialMemoryInMB\":305,\"UsedMemoryInMB\":40,\"CommittedMemoryInMB\":314,\"MaxMemoryInMB\":-1,\"CumulativeGCTimeInSeconds\":0,\"CumulativeNumberOfGCCollections\":132},{\"name\":\"Heap Memory\",\"InitialMemoryInMB\":32,\"UsedMemoryInMB\":18,\"CommittedMemoryInMB\":34,\"MaxMemoryInMB\":256},{\"name\":\"Non-Heap Memory\",\"InitialMemoryInMB\":273,\"UsedMemoryInMB\":22,\"CommittedMemoryInMB\":280,\"MaxMemoryInMB\":-1},{\"name\":\"Garbage Collection - scavenge\",\"CumulativeGCTimeInSeconds\":0,\"CumulativeNumberOfGCCollections\":131},{\"name\":\"Garbage Collection - global\",\"CumulativeGCTimeInSeconds\":0,\"CumulativeNumberOfGCCollections\":1}]}]}},\"event\":0}"

	var sds StatisticsDataStruct

	unmarshallError := json.Unmarshal([]byte(resourceStatisticsString), &sds)

	if unmarshallError != nil {
		t.Errorf("Error parsing json: %e", unmarshallError)
	}

	// s, _ := json.Marshal(sds)
	// t.Errorf("resource json=%s", string(s))

	for _, v := range sds.Data.ResourceStatistics.ResourceType {
		if v.Name == "JVM" {
			jvmData := NewJVMData(v.ResourceIdentifier)
			if jvmData.SummaryInitial != 305 {
				t.Errorf("Expected value=%d; actual=%d", 305, jvmData.SummaryInitial)
			}
			if jvmData.SummaryUsed != 40 {
				t.Errorf("Expected value=%d; actual=%d", 40, jvmData.SummaryUsed)
			}
			if jvmData.SummaryCommitted != 314 {
				t.Errorf("Expected value=%d; actual=%d", 314, jvmData.SummaryCommitted)
			}
			if jvmData.SummaryMax != -1 {
				t.Errorf("Expected value=%d; actual=%d", -1, jvmData.SummaryMax)
			}
			if jvmData.SummaryGCTime != 0 {
				t.Errorf("Expected value=%d; actual=%d", 0, jvmData.SummaryGCTime)
			}
			if jvmData.SummaryGCCount != 132 {
				t.Errorf("Expected value=%d; actual=%d", 132, jvmData.SummaryGCCount)
			}
			if jvmData.HeapInitial != 32 {
				t.Errorf("Expected value=%d; actual=%d", 32, jvmData.HeapInitial)
			}
			if jvmData.HeapUsed != 18 {
				t.Errorf("Expected value=%d; actual=%d", 18, jvmData.HeapUsed)
			}
			if jvmData.HeapCommitted != 34 {
				t.Errorf("Expected value=%d; actual=%d", 34, jvmData.HeapCommitted)
			}
			if jvmData.HeapMax != 256 {
				t.Errorf("Expected value=%d; actual=%d", 256, jvmData.HeapMax)
			}
			if jvmData.NativeInitial != 273 {
				t.Errorf("Expected value=%d; actual=%d", 273, jvmData.NativeInitial)
			}
			if jvmData.NativeUsed != 22 {
				t.Errorf("Expected value=%d; actual=%d", 22, jvmData.NativeUsed)
			}
			if jvmData.NativeCommitted != 280 {
				t.Errorf("Expected value=%d; actual=%d", 280, jvmData.NativeCommitted)
			}
			if jvmData.NativeMax != -1 {
				t.Errorf("Expected value=%d; actual=%d", -1, jvmData.NativeMax)
			}
			if jvmData.ScavengerGCTime != 0 {
				t.Errorf("Expected value=%d; actual=%d", 0, jvmData.ScavengerGCTime)
			}
			if jvmData.ScavengerGCCount != 131 {
				t.Errorf("Expected value=%d; actual=%d", 131, jvmData.ScavengerGCCount)
			}
			if jvmData.GlobalGCTime != 0 {
				t.Errorf("Expected value=%d; actual=%d", 0, jvmData.GlobalGCTime)
			}
			if jvmData.GlobalGCCount != 1 {
				t.Errorf("Expected value=%d; actual=%d", 1, jvmData.GlobalGCCount)
			}
		}
	}

	//TODO: Doesn't seem to give code coverage of scavenger of global gc data - are these values being picked up?
}
