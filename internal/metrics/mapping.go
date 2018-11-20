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

const (
	ResourceStatisticsData      = 0
	AccountingAndStatisticsData = 2
)

type MetricType int

const (
	Total   MetricType = 0
	Minimum MetricType = 1
	Maximum MetricType = 2
	Current MetricType = 3
)

type MetricUnits int

const (
	Microseconds MetricUnits = 0
	Seconds      MetricUnits = 1
	Bytes        MetricUnits = 2
	MegaBytes    MetricUnits = 3
	Count        MetricUnits = 4
)

type MetricLevel int

const (
	MsgFlowLevel     MetricLevel = 0
	MsgFlowNodeLevel MetricLevel = 1
	Resource         MetricLevel = 2
)

type metricLookup struct {
	name        string
	description string
	enabled     bool
	metricType  MetricType
	metricUnits MetricUnits
	metricLevel MetricLevel
}

type ThreadStatisticsStruct struct {
}

type TerminalStatisticsStruct struct {
}

type NodesStatisticsStruct struct {
	Label                   string `json:"Label"`
	Type                    string `json:"Type"`
	TotalElapsedTime        int    `json:"TotalElapsedTime"`
	MaximumElapsedTime      int    `json:"MaximumElapsedTime"`
	MinimumElapsedTime      int    `json:"MinimumElapsedTime"`
	TotalCPUTime            int    `json:"TotalCPUTime"`
	MaximumCPUTime          int    `json:"MaximumCPUTime"`
	MinimumCPUTime          int    `json:"MinimumCPUTime"`
	CountOfInvocations      int    `json:"CountOfInvocations"`
	NumberOfInputTerminals  int    `json:"NumberOfInputTerminals"`
	NumberOfOutputTerminals int    `json:"NumberOfOutputTerminals"`
	TerminalStatistics      []TerminalStatisticsStruct
}

type MessageFlowStruct struct {
	BrokerLabel                                               string `json:"BrokerLabel"`
	BrokerUUID                                                string `json:"BrokerUUID"`
	ExecutionGroupName                                        string `json:"ExecutionGroupName"`
	ExecutionGroupUUID                                        string `json:"ExecutionGroupUUID"`
	MessageFlowName                                           string `json:"MessageFlowName"`
	ApplicationName                                           string `json:"ApplicationName"`
	StartDate                                                 string `json:"StartDate"`
	StartTime                                                 string `json:"StartTime"`
	GMTStartTime                                              string `json:"GMTStartTime"`
	EndDate                                                   string `json:"EndDate"`
	EndTime                                                   string `json:"EndTime"`
	GMTEndTime                                                string `json:"GMTEndTime"`
	TotalElapsedTime                                          int    `json:"TotalElapsedTime"`
	MaximumElapsedTime                                        int    `json:"MaximumElapsedTime"`
	MinimumElapsedTime                                        int    `json:"MinimumElapsedTime"`
	TotalCPUTime                                              int    `json:"TotalCPUTime"`
	MaximumCPUTime                                            int    `json:"MaximumCPUTime"`
	MinimumCPUTime                                            int    `json:"MinimumCPUTime"`
	CPUTimeWaitingForInputMessage                             int    `json:"CPUTimeWaitingForInputMessage"`
	ElapsedTimeWaitingForInputMessage                         int    `json:"ElapsedTimeWaitingForInputMessage"`
	TotalInputMessages                                        int    `json:"TotalInputMessages"`
	TotalSizeOfInputMessages                                  int    `json:"TotalSizeOfInputMessages"`
	MaximumSizeOfInputMessages                                int    `json:"MaximumSizeOfInputMessages"`
	MinimumSizeOfInputMessages                                int    `json:"MinimumSizeOfInputMessages"`
	NumberOfThreadsInPool                                     int    `json:"NumberOfThreadsInPool"`
	TimesMaximumNumberOfThreadsReached                        int    `json:"TimesMaximumNumberOfThreadsReached"`
	TotalNumberOfMQErrors                                     int    `json:"TotalNumberOfMQErrors"`
	TotalNumberOfMessagesWithErrors                           int    `json:"TotalNumberOfMessagesWithErrors"`
	TotalNumberOfErrorsProcessingMessages                     int    `json:"TotalNumberOfErrorsProcessingMessages"`
	TotalNumberOfTimeOutsWaitingForRepliesToAggregateMessages int    `json:"TotalNumberOfTimeOutsWaitingForRepliesToAggregateMessages"`
	TotalNumberOfCommits                                      int    `json:"TotalNumberOfCommits"`
	TotalNumberOfBackouts                                     int    `json:"TotalNumberOfBackouts"`
	AccountingOrigin                                          string `json:"AccountingOrigin"`
}

type DataStruct struct {
	WMQIStatisticsAccounting *WMQIStatisticsAccountingStruct `json:"WMQIStatisticsAccounting,omitempty"`
	ResourceStatistics       *ResourceStatisticsStruct       `json:"ResourceStatistics,omitempty"`
}

type WMQIStatisticsAccountingStruct struct {
	RecordType       string                   `json:"RecordType"`
	RecordCode       string                   `json:"RecordCode"`
	MessageFlow      MessageFlowStruct        `json:"MessageFlow"`
	NumberOfThreads  int                      `json:"NumberOfThreads"`
	ThreadStatistics []ThreadStatisticsStruct `json:"ThreadStatistics"`
	NumberOfNodes    int                      `json:"NumberOfNodes"`
	Nodes            []NodesStatisticsStruct  `json:"Nodes"`
}

type StatisticsDataStruct struct {
	Data  DataStruct `json:"data"`
	Event int        `json:"event"`
}

type ResourceStatisticsStruct struct {
	BrokerLabel         string               `json:"brokerLabel"`
	BrokerUUID          string               `json:"brokerUUID"`
	ExecutionGroupName  string               `json:"executionGroupName"`
	ExecutiongGroupUUID string               `json:"executionGroupUUID"`
	CollectionStartDate string               `json:"collectionStartDate"`
	CollectionStartTime string               `json:"collectionStartTime"`
	StartDate           string               `json:"startDate"`
	StartTime           string               `json:"startTime"`
	EndDate             string               `json:"endDate"`
	EndTime             string               `json:"endTime"`
	Timezone            string               `json:"timezone"`
	ResourceType        []ResourceTypeStruct `json:"ResourceType"`
}

type ResourceTypeStruct struct {
	Name               string                   `json:"name"`
	ResourceIdentifier []map[string]interface{} `json:"resourceIdentifier"`
}

type JvmDataStruct struct {
	SummaryInitial   int
	SummaryUsed      int
	SummaryCommitted int
	SummaryMax       int
	SummaryGCTime    int
	SummaryGCCount   int
	HeapInitial      int
	HeapUsed         int
	HeapCommitted    int
	HeapMax          int
	NativeInitial    int
	NativeUsed       int
	NativeCommitted  int
	NativeMax        int
	ScavengerGCTime  int
	ScavengerGCCount int
	GlobalGCTime     int
	GlobalGCCount    int
}

func NewJVMData(ma []map[string]interface{}) *JvmDataStruct {

	jvmData := JvmDataStruct{}

	for _, m := range ma {
		switch m["name"] {
		case "summary":
			jvmData.SummaryInitial = int(m["InitialMemoryInMB"].(float64))
			jvmData.SummaryUsed = int(m["UsedMemoryInMB"].(float64))
			jvmData.SummaryCommitted = int(m["CommittedMemoryInMB"].(float64))
			jvmData.SummaryMax = int(m["MaxMemoryInMB"].(float64))
			jvmData.SummaryGCTime = int(m["CumulativeGCTimeInSeconds"].(float64))
			jvmData.SummaryGCCount = int(m["CumulativeNumberOfGCCollections"].(float64))
		case "Heap Memory":
			jvmData.HeapInitial = int(m["InitialMemoryInMB"].(float64))
			jvmData.HeapUsed = int(m["UsedMemoryInMB"].(float64))
			jvmData.HeapCommitted = int(m["CommittedMemoryInMB"].(float64))
			jvmData.HeapMax = int(m["MaxMemoryInMB"].(float64))
		case "Non-Heap Memory":
			jvmData.NativeInitial = int(m["InitialMemoryInMB"].(float64))
			jvmData.NativeUsed = int(m["UsedMemoryInMB"].(float64))
			jvmData.NativeCommitted = int(m["CommittedMemoryInMB"].(float64))
			jvmData.NativeMax = int(m["MaxMemoryInMB"].(float64))
		case "Garbage Collection - scavenge":
			jvmData.ScavengerGCTime = int(m["CumulativeGCTimeInSeconds"].(float64))
			jvmData.ScavengerGCCount = int(m["CumulativeNumberOfGCCollections"].(float64))
		case "Garbage Collection - global":
			jvmData.GlobalGCTime = int(m["CumulativeGCTimeInSeconds"].(float64))
			jvmData.GlobalGCCount = int(m["CumulativeNumberOfGCCollections"].(float64))
		}
	}

	return &jvmData
}

//  generates metric names mapped from their description
func generateMetricNamesMap() (msgFlowMetricNamesMap, msgFlowNodeMetricNamesMap map[string]metricLookup) {

	msgFlowMetricNamesMap = map[string]metricLookup{
		"MsgFlow/TotalElapsedTime":                                          metricLookup{"msgflow_elapsed_time_seconds_total", "Total elapsed time spent processing messages by the message flow", true, Total, Microseconds, MsgFlowLevel},
		"MsgFlow/MaximumElapsedTime":                                        metricLookup{"msgflow_elapsed_time_seconds_max", "Maximum elapsed time spent processing a message by the message flow", true, Maximum, Microseconds, MsgFlowLevel},
		"MsgFlow/MinimumElapsedTime":                                        metricLookup{"msgflow_elapsed_time_seconds_min", "Minimum elapsed time spent processing a message by the message flow", true, Minimum, Microseconds, MsgFlowLevel},
		"MsgFlow/TotalCpuTime":                                              metricLookup{"msgflow_cpu_time_seconds_total", "Total CPU time spent processing messages by the message flow", true, Total, Microseconds, MsgFlowLevel},
		"MsgFlow/MaximumCpuTime":                                            metricLookup{"msgflow_cpu_time_seconds_max", "Maximum CPU time spent processing a message by the message flow", true, Maximum, Microseconds, MsgFlowLevel},
		"MsgFlow/MinimumCpuTime":                                            metricLookup{"msgflow_cpu_time_seconds_min", "Minimum CPU time spent processing a message by the message flow", true, Minimum, Microseconds, MsgFlowLevel},
		"MsgFlow/TotalSizeOfInputMessages":                                  metricLookup{"msgflow_messages_bytes_total", "Total size of messages processed by the message flow", true, Total, Bytes, MsgFlowLevel},
		"MsgFlow/MaximumSizeOfInputMessages":                                metricLookup{"msgflow_messages_bytes_max", "Maximum size of message processed by the message flow", true, Maximum, Bytes, MsgFlowLevel},
		"MsgFlow/MinimumSizeOfInputMessages":                                metricLookup{"msgflow_messages_bytes_min", "Minimum size of message processed by the message flow", true, Minimum, Bytes, MsgFlowLevel},
		"MsgFlow/TotalInputMessages":                                        metricLookup{"msgflow_messages_total", "Total number of messages processed by the message flow", true, Total, Count, MsgFlowLevel},
		"MsgFlow/TotalCPUTimeWaiting":                                       metricLookup{"msgflow_cpu_time_waiting_seconds_total", "Total CPU time spent waiting for input messages by the message flow", true, Total, Microseconds, MsgFlowLevel},
		"MsgFlow/TotalElapsedTimeWaiting":                                   metricLookup{"msgflow_elapsed_time_waiting_seconds_total", "Total elapsed time spent waiting for input messages by the message flow", true, Total, Microseconds, MsgFlowLevel},
		"MsgFlow/NumberOfThreadsInPool":                                     metricLookup{"msgflow_threads_total", "Number of threads in the pool for the message flow", true, Current, Count, MsgFlowLevel},
		"MsgFlow/TimesMaximumNumberOfThreadsReached":                        metricLookup{"msgflow_threads_reached_maximum_total", "Number of times that maximum number of threads in the pool for the message flow was reached", true, Total, Count, MsgFlowLevel},
		"MsgFlow/TotalNumberOfMQErrors":                                     metricLookup{"msgflow_mq_errors_total", "Total number of MQ errors in the message flow", true, Total, Count, MsgFlowLevel},
		"MsgFlow/TotalNumberOfMessagesWithErrors":                           metricLookup{"msgflow_messages_with_error_total", "Total number of messages processed by the message flow that had errors", true, Total, Count, MsgFlowLevel},
		"MsgFlow/TotalNumberOfErrorsProcessingMessages":                     metricLookup{"msgflow_errors_total", "Total number of errors processing messages by the message flow", true, Total, Count, MsgFlowLevel},
		"MsgFlow/TotalNumberOfTimeOutsWaitingForRepliesToAggregateMessages": metricLookup{"msgflow_aggregation_timeouts_total", "Total number of timeouts waiting for replies to Aggregate messages", true, Total, Count, MsgFlowLevel},
		"MsgFlow/TotalNumberOfCommits":                                      metricLookup{"msgflow_commits_total", "Total number of commits by the message flow", true, Total, Count, MsgFlowLevel},
		"MsgFlow/TotalNumberOfBackouts":                                     metricLookup{"msgflow_backouts_total", "Total number of backouts by the message flow", true, Total, Count, MsgFlowLevel},
	}

	msgFlowNodeMetricNamesMap = map[string]metricLookup{
		"MsgFlowNode/TotalElapsedTime":   metricLookup{"msgflownode_elapsed_time_seconds_total", "Total elapsed time spent processing messages by the message flow node", true, Total, Microseconds, MsgFlowNodeLevel},
		"MsgFlowNode/MaximumElapsedTime": metricLookup{"msgflownode_elapsed_time_seconds_max", "Maximum elapsed time spent processing a message by the message flow node", true, Maximum, Microseconds, MsgFlowNodeLevel},
		"MsgFlowNode/MinimumElapsedTime": metricLookup{"msgflownode_elapsed_time_seconds_min", "Minimum elapsed time spent processing a message by the message flow node", true, Minimum, Microseconds, MsgFlowNodeLevel},
		"MsgFlowNode/TotalCpuTime":       metricLookup{"msgflownode_cpu_time_seconds_total", "Total CPU time spent processing messages by the message flow node", true, Total, Microseconds, MsgFlowNodeLevel},
		"MsgFlowNode/MaximumCpuTime":     metricLookup{"msgflownode_cpu_time_seconds_max", "Maximum CPU time spent processing a message by the message flow node", true, Maximum, Microseconds, MsgFlowNodeLevel},
		"MsgFlowNode/MinimumCpuTime":     metricLookup{"msgflownode_cpu_time_seconds_min", "Minimum CPU time spent processing a message by the message flow node", true, Minimum, Microseconds, MsgFlowNodeLevel},
		"MsgFlowNode/TotalInvocations":   metricLookup{"msgflownode_messages_total", "Total number of messages processed by the message flow node", true, Total, Count, MsgFlowNodeLevel},
		"MsgFlowNode/InputTerminals":     metricLookup{"msgflownode_input_terminals_total", "Total number of input terminals on the message flow node", true, Current, Count, MsgFlowNodeLevel},
		"MsgFlowNode/OutputTerminals":    metricLookup{"msgflownode_output_terminals_total", "Total number of output terminals on the message flow node", true, Current, Count, MsgFlowNodeLevel},
	}

	return
}

//  generates metric names mapped from their description
func generateResourceMetricNamesMap() (jvmMetricNamesMap map[string]metricLookup) {

	jvmMetricNamesMap = map[string]metricLookup{
		"JVM/Summary/InitialMemoryInMB":                   metricLookup{"jvm_summary_initial_memory_bytes", "Initial memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Summary/UsedMemoryInMB":                      metricLookup{"jvm_summary_used_memory_bytes", "Used memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Summary/CommittedMemoryInMB":                 metricLookup{"jvm_summary_committed_memory_bytes", "Committed memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Summary/MaxMemoryInMB":                       metricLookup{"jvm_summary_max_memory_bytes", "Committed memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Summary/CumulativeGCTimeInSeconds":           metricLookup{"jvm_summary_gcs_elapsed_time_seconds_total", "Total time spent in GCs for the JVM", true, Current, Seconds, Resource},
		"JVM/Summary/CumulativeNumberOfGCCollections":     metricLookup{"jvm_summary_gcs_total", "Total number of GCs for the JVM", true, Current, Count, Resource},
		"JVM/Heap/InitialMemoryInMB":                      metricLookup{"jvm_heap_initial_memory_bytes", "Initial heap memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Heap/UsedMemoryInMB":                         metricLookup{"jvm_heap_used_memory_bytes", "Used heap memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Heap/CommittedMemoryInMB":                    metricLookup{"jvm_heap_committed_memory_bytes", "Committed heap memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Heap/MaxMemoryInMB":                          metricLookup{"jvm_heap_max_memory_bytes", "Committed heap memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Native/InitialMemoryInMB":                    metricLookup{"jvm_native_initial_memory_bytes", "Initial native memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Native/UsedMemoryInMB":                       metricLookup{"jvm_native_used_memory_bytes", "Used native memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Native/CommittedMemoryInMB":                  metricLookup{"jvm_native_committed_memory_bytes", "Committed native memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/Native/MaxMemoryInMB":                        metricLookup{"jvm_native_max_memory_bytes", "Committed native memory for the JVM", true, Current, MegaBytes, Resource},
		"JVM/ScavengerGC/CumulativeGCTimeInSeconds":       metricLookup{"jvm_scavenger_gcs_elapsed_time_seconds_total", "Total time spent in scavenger GCs for the JVM", true, Current, Seconds, Resource},
		"JVM/ScavengerGC/CumulativeNumberOfGCCollections": metricLookup{"jvm_scavenger_gcs_total", "Total number of scavenger GCs for the JVM", true, Current, Count, Resource},
		"JVM/GlobalGC/CumulativeGCTimeInSeconds":          metricLookup{"jvm_global_gcs_elapsed_time_seconds_total", "Total time spent in global GCs for the JVM", true, Current, Seconds, Resource},
		"JVM/GlobalGC/CumulativeNumberOfGCCollections":    metricLookup{"jvm_global_gcs_total", "Total number of global GCs for the JVM", true, Current, Count, Resource},
	}

	return
}
