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
    "crypto/tls"
    "crypto/x509"
    "errors"
    "io/ioutil"
    "flag"
    "fmt"
    "math"
    "net/url"
    "os"
    "sync"

    "github.com/ot4i/ace-docker/internal/logger"

    "github.com/gorilla/websocket"
    "github.com/prometheus/client_golang/prometheus"
)

var (
    addr              = flag.String("addr", "localhost:7600", "http service address")
    startChannel      = make(chan bool)
    stopChannel       = make(chan bool, 2)
    stopping          = false
    requestChannel    = make(chan bool)
    responseChannel   = make(chan *MetricsMap)
    statisticsChannel = make(chan StatisticsDataStruct, 10) // Block on writing to channel if we already have 10 queued so we don't retrieve any more
)

type MetricsMap struct {
    sync.Mutex
    internal map[string]*metricData
}

func NewMetricsMap() *MetricsMap {
    return &MetricsMap{
        internal: make(map[string]*metricData),
    }
}

type Metric struct {
    labels prometheus.Labels
    value  float64
}

type metricData struct {
    name        string
    description string
    values      map[string]*Metric
    metricType  MetricType
    metricUnits MetricUnits
    metricLevel MetricLevel
}

/*
Normalise returns a float64 representation of the metric value normalised to a base metric type (seconds, bytes, etc.)
*/
func (md *metricData) Normalise(value int) float64 {
    f := float64(value)

    if f < 0 {
        f = 0
    }

    // Convert microseconds to seconds
    if md.metricUnits == Microseconds {
        f = f / 1000000
    }

    // Convert megabytes to bytes
    if md.metricUnits == MegaBytes {
        f = f * 1024 * 1024
    }

    return f
}

func ReadStatistics(log *logger.Logger) {
    // Check if the admin server is secure so we know whether to connect with wss or ws
    aceAdminServerSecurity := os.Getenv("ACE_ADMIN_SERVER_SECURITY")
    if aceAdminServerSecurity == "" {
        log.Printf("Can't tell if ace admin server security is enabled defaulting to false")
        aceAdminServerSecurity = "false"
    } else {
        log.Printf("ACE_ADMIN_SERVER_SECURITY is %s", aceAdminServerSecurity)
    }

    var firstConnect = true

    for {
        if stopping {
            // Stopping will trigger a read error on the c.ReadJSON call and re-entry into this loop,
            // but we want to exit this function when that happens
            return
        }

        var c *websocket.Conn
        var dialError error

        // Use wss with TLS if using the admin server is secured
        if aceAdminServerSecurity == "true" {
            adminServerCACert :=  os.Getenv("ACE_ADMIN_SERVER_CA")
            log.Printf("Using CA Certificate file %s", adminServerCACert)
            caCert, err := ioutil.ReadFile(adminServerCACert)
            if err != nil {
                log.Errorf("Error reading CA Certificate %s", err)
                return
            }
            caCertPool := x509.NewCertPool()
            ok := caCertPool.AppendCertsFromPEM(caCert)
            if !ok {
                log.Errorf("failed to parse root CA Certificate")
            }

            // Read the key/ cert pair to create tls certificate
            adminServerCert := os.Getenv("ACE_ADMIN_SERVER_CERT")
            adminServerKey := os.Getenv("ACE_ADMIN_SERVER_KEY")
            adminServerCerts, err := tls.LoadX509KeyPair(adminServerCert, adminServerKey)
            if err != nil {
                if ( adminServerCert != "" && adminServerKey != "" ) {
                    log.Errorf("Error reading TLS Certificates: %s", err)
                    return
                }
            } else {
                log.Printf("Using provided cert and key for mutual auth")
            }

            aceAdminServerName := os.Getenv("ACE_ADMIN_SERVER_NAME")
            if aceAdminServerName == "" {
                log.Printf("No ace admin server name available")
                return
            } else {
                log.Printf("ACE_ADMIN_SERVER_NAME is %s", aceAdminServerName)
            }

            u := url.URL{Scheme: "wss", Host: *addr, Path: "/"}
            log.Printf("Connecting to %s for statistics gathering", u.String())
            d := websocket.Dialer{
                TLSClientConfig: &tls.Config{
                    RootCAs: caCertPool,
                    Certificates: []tls.Certificate{adminServerCerts},
                    ServerName: aceAdminServerName,
                },
            }
            c, _, dialError = d.Dial(u.String(), nil)
        } else {
            u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
            log.Printf("Connecting to %s for statistics", u.String())
            c, _, dialError = websocket.DefaultDialer.Dial(u.String(), nil)
        }

        if dialError == nil {
            if firstConnect {
                firstConnect = false
                startChannel <- true
            }

            defer c.Close()

            // Loop reading from websocket and put messages on the statistics statisticsChannel
            // End the loop and reconnect if there is an error reading from the websocket
            var readError error
            for readError == nil {
                var m StatisticsDataStruct

                readError = c.ReadJSON(&m)
                if readError == nil {
                    statisticsChannel <- m
                }
            }
        } else {
            log.Errorf("Error calling ace admin server webservice endpoint %s", dialError)
        }
    }
}

// processMetrics processes publications of metric data and handles describe/collect/stop requests
func processMetrics(log *logger.Logger, serverName string) {
    log.Println("Processing metrics...")

    metrics := initialiseMetrics(log)

    go ReadStatistics(log)

    // Handle update/describe/collect/stop requests
    for {
        select {
        case m := <-statisticsChannel:
            newMetrics, parseError := parseMetrics(log, &m)

            if parseError != nil {
                log.Println("Parse Error:", parseError)
            } else {
                updateMetrics(log, metrics, newMetrics)
            }
        case <-requestChannel:
            responseChannel <- metrics
        case <-stopChannel:
            log.Println("Stopping metrics gathering")
            stopping = true
            return
        }
    }
}

// initialiseMetrics sets initial details for all available metrics
func initialiseMetrics(log *logger.Logger) *MetricsMap {

    metrics := NewMetricsMap()
    msgFlowMetricNamesMap, msgFlowNodeMetricNamesMap := generateMetricNamesMap()

    for k, v := range msgFlowMetricNamesMap {
        if v.enabled {
            // Set metric details
            metric := metricData{
                name:        v.name,
                description: v.description,
                metricType:  v.metricType,
                metricUnits: v.metricUnits,
                metricLevel: v.metricLevel,
            }
            metric.values = make(map[string]*Metric)

            // Add metric
            metrics.internal[k] = &metric
        }
    }

    for k, v := range msgFlowNodeMetricNamesMap {
        if v.enabled {
            // Set metric details
            metric := metricData{
                name:        v.name,
                description: v.description,
                metricType:  v.metricType,
                metricUnits: v.metricUnits,
                metricLevel: v.metricLevel,
            }
            metric.values = make(map[string]*Metric)

            // Add metric
            metrics.internal[k] = &metric
        }
    }

    jvmResourceMetricNamesMap := generateResourceMetricNamesMap()

    for k, v := range jvmResourceMetricNamesMap {
        if v.enabled {
            // Set metric details
            metric := metricData{
                name:        v.name,
                description: v.description,
                metricType:  v.metricType,
                metricUnits: v.metricUnits,
                metricLevel: v.metricLevel,
            }
            metric.values = make(map[string]*Metric)

            // Add metric
            metrics.internal[k] = &metric
        }
    }

    return metrics
}

func parseMetrics(log *logger.Logger, m *StatisticsDataStruct) (*MetricsMap, error) {
    if m.Event == ResourceStatisticsData {
        return parseResourceMetrics(log, m)
    } else if m.Event == AccountingAndStatisticsData {
        return parseAccountingMetrics(log, m)
    } else {
        return nil, fmt.Errorf("Unable to parse data with event: %d", m.Event)
    }
}

func parseAccountingMetrics(log *logger.Logger, m *StatisticsDataStruct) (*MetricsMap, error) {
    parsedMetrics := NewMetricsMap()

    msgFlowMetricNamesMap, msgFlowNodeMetricNamesMap := generateMetricNamesMap()

    accountingOrigin := m.Data.WMQIStatisticsAccounting.MessageFlow.AccountingOrigin
    serverName := m.Data.WMQIStatisticsAccounting.MessageFlow.ExecutionGroupName
    applicationName := m.Data.WMQIStatisticsAccounting.MessageFlow.ApplicationName
    msgflowName := m.Data.WMQIStatisticsAccounting.MessageFlow.MessageFlowName

    if msgflowName == "" {
        err := errors.New("parse error - no message flow name in statistics")
        return parsedMetrics, err
    }

    flowValuesMap := map[string]int{
        "MsgFlow/TotalElapsedTime":                                          m.Data.WMQIStatisticsAccounting.MessageFlow.TotalElapsedTime,
        "MsgFlow/MaximumElapsedTime":                                        m.Data.WMQIStatisticsAccounting.MessageFlow.MaximumElapsedTime,
        "MsgFlow/MinimumElapsedTime":                                        m.Data.WMQIStatisticsAccounting.MessageFlow.MinimumElapsedTime,
        "MsgFlow/TotalCpuTime":                                              m.Data.WMQIStatisticsAccounting.MessageFlow.TotalCPUTime,
        "MsgFlow/MaximumCpuTime":                                            m.Data.WMQIStatisticsAccounting.MessageFlow.MaximumCPUTime,
        "MsgFlow/MinimumCpuTime":                                            m.Data.WMQIStatisticsAccounting.MessageFlow.MinimumCPUTime,
        "MsgFlow/TotalSizeOfInputMessages":                                  m.Data.WMQIStatisticsAccounting.MessageFlow.TotalSizeOfInputMessages,
        "MsgFlow/MaximumSizeOfInputMessages":                                m.Data.WMQIStatisticsAccounting.MessageFlow.MaximumSizeOfInputMessages,
        "MsgFlow/MinimumSizeOfInputMessages":                                m.Data.WMQIStatisticsAccounting.MessageFlow.MinimumSizeOfInputMessages,
        "MsgFlow/TotalInputMessages":                                        m.Data.WMQIStatisticsAccounting.MessageFlow.TotalInputMessages,
        "MsgFlow/TotalCPUTimeWaiting":                                       m.Data.WMQIStatisticsAccounting.MessageFlow.CPUTimeWaitingForInputMessage,
        "MsgFlow/TotalElapsedTimeWaiting":                                   m.Data.WMQIStatisticsAccounting.MessageFlow.ElapsedTimeWaitingForInputMessage,
        "MsgFlow/NumberOfThreadsInPool":                                     m.Data.WMQIStatisticsAccounting.MessageFlow.NumberOfThreadsInPool,
        "MsgFlow/TimesMaximumNumberOfThreadsReached":                        m.Data.WMQIStatisticsAccounting.MessageFlow.TimesMaximumNumberOfThreadsReached,
        "MsgFlow/TotalNumberOfMQErrors":                                     m.Data.WMQIStatisticsAccounting.MessageFlow.TotalNumberOfMQErrors,
        "MsgFlow/TotalNumberOfMessagesWithErrors":                           m.Data.WMQIStatisticsAccounting.MessageFlow.TotalNumberOfMessagesWithErrors,
        "MsgFlow/TotalNumberOfErrorsProcessingMessages":                     m.Data.WMQIStatisticsAccounting.MessageFlow.TotalNumberOfErrorsProcessingMessages,
        "MsgFlow/TotalNumberOfTimeOutsWaitingForRepliesToAggregateMessages": m.Data.WMQIStatisticsAccounting.MessageFlow.TotalNumberOfTimeOutsWaitingForRepliesToAggregateMessages,
        "MsgFlow/TotalNumberOfCommits":                                      m.Data.WMQIStatisticsAccounting.MessageFlow.TotalNumberOfCommits,
        "MsgFlow/TotalNumberOfBackouts":                                     m.Data.WMQIStatisticsAccounting.MessageFlow.TotalNumberOfBackouts,
    }

    /*
      Process flow level accounting and statistics data
    */
    for k, v := range flowValuesMap {
        metricDesc := msgFlowMetricNamesMap[k]
        if metricDesc.enabled {
            metric := metricData{
                name:        metricDesc.name,
                description: metricDesc.description,
                metricType:  metricDesc.metricType,
                metricUnits: metricDesc.metricUnits,
                metricLevel: metricDesc.metricLevel,
            }
            metric.values = make(map[string]*Metric)
            metric.values[accountingOrigin+"_"+applicationName+"_"+msgflowName] = &Metric{labels: prometheus.Labels{msgflowPrefix: msgflowName, serverLabel: serverName, applicationLabel: applicationName, originLabel: accountingOrigin}, value: metric.Normalise(v)}
            parsedMetrics.internal[k] = &metric
        }
    }

    /*
        Process node level accounting and statistics data
    */
    for k, v := range msgFlowNodeMetricNamesMap {

        if v.enabled {
            metric := metricData{
                name:        v.name,
                description: v.description,
                metricType:  v.metricType,
                metricUnits: v.metricUnits,
                metricLevel: v.metricLevel,
            }
            metric.values = make(map[string]*Metric)

            for _, node := range m.Data.WMQIStatisticsAccounting.Nodes {
                nodeValuesMap := map[string]int{
                    "MsgFlowNode/TotalElapsedTime":   node.TotalElapsedTime,
                    "MsgFlowNode/MaximumElapsedTime": node.MaximumElapsedTime,
                    "MsgFlowNode/MinimumElapsedTime": node.MinimumElapsedTime,
                    "MsgFlowNode/TotalCpuTime":       node.TotalCPUTime,
                    "MsgFlowNode/MaximumCpuTime":     node.MaximumCPUTime,
                    "MsgFlowNode/MinimumCpuTime":     node.MinimumCPUTime,
                    "MsgFlowNode/TotalInvocations":   node.CountOfInvocations,
                    "MsgFlowNode/InputTerminals":     node.NumberOfInputTerminals,
                    "MsgFlowNode/OutputTerminals":    node.NumberOfOutputTerminals,
                }
                msgflownodeName := node.Label
                msgflownodeType := node.Type

                metric.values[accountingOrigin+"_"+applicationName+"_"+msgflowName+"_"+msgflownodeName] = &Metric{labels: prometheus.Labels{msgflownodeLabel: msgflownodeName, msgflownodeTypeLabel: msgflownodeType, msgflowLabel: msgflowName, serverLabel: serverName, applicationLabel: applicationName, originLabel: accountingOrigin}, value: metric.Normalise(nodeValuesMap[k])}
            }
            parsedMetrics.internal[k] = &metric
        }
    }

    return parsedMetrics, nil
}

func parseResourceMetrics(log *logger.Logger, m *StatisticsDataStruct) (*MetricsMap, error) {
    parsedResourceMetrics := NewMetricsMap()

    serverName := m.Data.ResourceStatistics.ExecutionGroupName

    for _, v := range m.Data.ResourceStatistics.ResourceType {
        switch v.Name {
        case "JVM":
            jvmData := NewJVMData(v.ResourceIdentifier)

            jvmResourceMetricNamesMap := generateResourceMetricNamesMap()

            jvmValuesMap := map[string]int{
                "JVM/Summary/InitialMemoryInMB":                   jvmData.SummaryInitial,
                "JVM/Summary/UsedMemoryInMB":                      jvmData.SummaryUsed,
                "JVM/Summary/CommittedMemoryInMB":                 jvmData.SummaryCommitted,
                "JVM/Summary/MaxMemoryInMB":                       jvmData.SummaryMax,
                "JVM/Summary/CumulativeGCTimeInSeconds":           jvmData.SummaryGCTime,
                "JVM/Summary/CumulativeNumberOfGCCollections":     jvmData.SummaryGCCount,
                "JVM/Heap/InitialMemoryInMB":                      jvmData.HeapInitial,
                "JVM/Heap/UsedMemoryInMB":                         jvmData.HeapUsed,
                "JVM/Heap/CommittedMemoryInMB":                    jvmData.HeapCommitted,
                "JVM/Heap/MaxMemoryInMB":                          jvmData.HeapMax,
                "JVM/Native/InitialMemoryInMB":                    jvmData.NativeInitial,
                "JVM/Native/UsedMemoryInMB":                       jvmData.NativeUsed,
                "JVM/Native/CommittedMemoryInMB":                  jvmData.NativeCommitted,
                "JVM/Native/MaxMemoryInMB":                        jvmData.NativeMax,
                "JVM/ScavengerGC/CumulativeGCTimeInSeconds":       jvmData.ScavengerGCTime,
                "JVM/ScavengerGC/CumulativeNumberOfGCCollections": jvmData.ScavengerGCCount,
                "JVM/GlobalGC/CumulativeGCTimeInSeconds":          jvmData.GlobalGCTime,
                "JVM/GlobalGC/CumulativeNumberOfGCCollections":    jvmData.GlobalGCCount,
            }

            for metricKey, metricDesc := range jvmResourceMetricNamesMap {
                if metricDesc.enabled {
                    metric := metricData{
                        name:        metricDesc.name,
                        description: metricDesc.description,
                        metricType:  metricDesc.metricType,
                        metricUnits: metricDesc.metricUnits,
                        metricLevel: metricDesc.metricLevel,
                    }
                    metric.values = make(map[string]*Metric)
                    metric.values[metricKey] = &Metric{labels: prometheus.Labels{serverLabel: serverName}, value: metric.Normalise(jvmValuesMap[metricKey])}
                    parsedResourceMetrics.internal[metricKey] = &metric
                }
            }
        default:
            //TODO: Support other resource statistic types
        }
    }

    return parsedResourceMetrics, nil
}

// updateMetrics updates values for all available metrics
func updateMetrics(log *logger.Logger, mm1 *MetricsMap, mm2 *MetricsMap) {
    mm1.Lock()
    mm2.Lock()
    defer mm1.Unlock()
    defer mm2.Unlock()

    for k, md2 := range mm2.internal {
        if md1, ok := mm1.internal[k]; ok {
            //Iterate over the labels
            for l, m2 := range md2.values {
                if m1, ok := md1.values[l]; ok {
                    switch md1.metricType {
                    case Total:
                        md1.values[l].value = m1.value + m2.value
                    case Maximum:
                        md1.values[l].value = math.Max(m1.value, m2.value)
                    case Minimum:
                        md1.values[l].value = math.Min(m1.value, m2.value)
                    case Current:
                        md1.values[l].value = m2.value
                    default:
                        log.Printf("Should not reach here - only a set enumeration of metric types. %d is unknown...", md1.metricType)
                    }
                } else {
                    md1.values[l] = m2
                }
            }
        } else {
            mm1.internal[k] = mm2.internal[k]
        }
    }
}
