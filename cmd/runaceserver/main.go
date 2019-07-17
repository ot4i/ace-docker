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

// runaceserver initializes, creates and starts a queue manager, as PID 1 in a container
package main

import (
	"errors"
	"os"

	"github.com/ot4i/ace-docker/internal/command"
	"github.com/ot4i/ace-docker/internal/metrics"
	"github.com/ot4i/ace-docker/internal/name"
	"github.com/ot4i/ace-docker/internal/qmgr"
)

func doMain() error {
	useQmgr := qmgr.UseQueueManager()
	var qmgrProcess command.BackgroundCmd
	var integrationServerProcess command.BackgroundCmd

	name, nameErr := name.GetIntegrationServerName()
	err := configureLogger(name)
	if err != nil {
		logTermination(err)
		return err
	}
	if nameErr != nil {
		logTermination(err)
		return err
	}

	accepted, err := checkLicense()
	if err != nil {
		logTerminationf("Error checking license acceptance: %v", err)
		return err
	}
	if !accepted {
		err = errors.New("License not accepted")
		logTermination(err)
		return err
	}

	performShutdown := func() {
		metrics.StopMetricsGathering()

		log.Print("Stopping Integration Server")
		stopIntegrationServer(integrationServerProcess)
		log.Print("Integration Server stopped")

		log.Print("Stopping Queue Mgr")
		qmgr.StopQueueManager(qmgrProcess)
		log.Print("Queue Mgr stopped")

		log.Print("Shutdown complete")
	}

	// Start signal handler
	signalControl := signalHandler(performShutdown)

	// Print out versioning information
	logVersionInfo()

	if useQmgr {

		log.Println("Starting MQ Initialisation")
		err = qmgr.InitializeMQ()
		if err != nil {
			logTermination(err)
			performShutdown()
			return err
		}

		log.Println("Starting queue manager")
		qmgrProcess = qmgr.StartQueueManager(log)
		if qmgrProcess.ReturnError != nil {
			logTermination(qmgrProcess.ReturnError)
			return qmgrProcess.ReturnError
		}

		log.Println("Waiting for queue manager to be ready")
		err = qmgr.WaitForQueueManager(log)
		if err != nil {
			logTermination(err)
			performShutdown()
			return err
		}
		log.Println("Queue Manager is ready")

		err = createSystemQueues()
		if err != nil {
			logTermination(err)
			performShutdown()
			return err
		}
	}

	log.Println("Checking for valid working directory")
	err = createWorkDir()
	if err != nil {
		logTermination(err)
		performShutdown()
		return err
	}

	err = initialIntegrationServerConfig()
	if err != nil {
		logTermination(err)
		performShutdown()
		return err
	}

	log.Println("Starting integration server")
	integrationServerProcess = startIntegrationServer()
	if integrationServerProcess.ReturnError != nil {
		logTermination(integrationServerProcess.ReturnError)
		return integrationServerProcess.ReturnError
	}

	log.Println("Waiting for integration server to be ready")
	err = waitForIntegrationServer()
	if err != nil {
		logTermination(err)
		performShutdown()
		return err
	}
	log.Println("Integration server is ready")

	enableMetrics := os.Getenv("ACE_ENABLE_METRICS")
	if enableMetrics == "true" || enableMetrics == "1" {
		go metrics.GatherMetrics(name, log)
	} else {
		log.Println("Metrics are disabled")
	}

	// Start reaping zombies from now on.
	// Start this here, so that we don't reap any sub-processes created
	// by this process (e.g. for crtmqm or strmqm)
	signalControl <- startReaping
	// Reap zombies now, just in case we've already got some
	signalControl <- reapNow
	// Wait for terminate signal
	<-signalControl

	return nil
}

var osExit = os.Exit

func main() {
	err := doMain()
	if err != nil {
		osExit(1)
	}
}
