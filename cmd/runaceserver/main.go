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
	"io/ioutil"
	"os"

	"github.com/ot4i/ace-docker/common/designer"
	"github.com/ot4i/ace-docker/internal/command"
	"github.com/ot4i/ace-docker/internal/configuration"
	iscommandsapi "github.com/ot4i/ace-docker/internal/isCommandsApi"
	"github.com/ot4i/ace-docker/internal/metrics"
	"github.com/ot4i/ace-docker/internal/name"
	"github.com/ot4i/ace-docker/internal/qmgr"
	"github.com/ot4i/ace-docker/internal/trace"
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

		// Stop watching the Force Flows Secret if the watcher has been created
		if watcher != nil {
			log.Print("Stopping watching the Force Flows HTTPS secret")
			watcher.Close()
		}

		log.Print("Stopping Integration Server")
		stopIntegrationServer(integrationServerProcess)
		log.Print("Integration Server stopped")

		log.Print("Stopping Queue Mgr")
		qmgr.StopQueueManager(qmgrProcess)
		log.Print("Queue Mgr stopped")

		checkLogs()

		iscommandsapi.StopCommandsAPIServer()
		log.Print("Shutdown complete")
	}

	restartIntegrationServer := func() error {
		err := ioutil.WriteFile("/tmp/integration_server_restart.timestamp", []byte(""), 0755)

		if err != nil {
			log.Print("RestartIntegrationServer - Creating restart file failed")
			return err
		}

		log.Print("RestartIntegrationServer - Stopping integration server")
		stopIntegrationServer(integrationServerProcess)
		log.Println("RestartIntegrationServer - Starting integration server")

		integrationServerProcess = startIntegrationServer()
		err = integrationServerProcess.ReturnError

		if integrationServerProcess.ReturnError == nil {
			log.Println("RestartIntegrationServer - Waiting for integration server")
			err = waitForIntegrationServer()
		}

		if err != nil {
			logTermination(err)
			performShutdown()
			return err
		}

		log.Println("RestartIntegrationServer - Integration server is ready")

		return nil
	}

	// Start signal handler
	signalControl := signalHandler(performShutdown)

	// Print out versioning information
	logVersionInfo()

	runOnly := os.Getenv("ACE_RUN_ONLY")
	if runOnly == "true" || runOnly == "1" {
			log.Println("Run selected so skipping setup")
	} else {
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

		// Note: this will do nothing if there are no crs set in the environment
		err = configuration.SetupConfigurationsFiles(log, "/home/aceuser")
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

		log.Println("Validating flows deployed to the integration server before starting")
		licenseToggles, err := designer.GetLicenseTogglesFromEnvironmentVariables()
		if err != nil {
			logTermination(err)
			performShutdown()
			return err
		}
		designer.InitialiseLicenseToggles(licenseToggles)

		err = designer.ValidateFlows(log, "/home/aceuser")
		if err != nil {
			logTermination(err)
			performShutdown()
			return err
		}

		// Apply any WorkdirOverrides provided
		err = applyWorkdirOverrides()
		if err != nil {
			logTermination(err)
			performShutdown()
			return err
		}
	}

	setupOnly := os.Getenv("ACE_SETUP_ONLY")
	if setupOnly == "true" || setupOnly == "1" {
		log.Println("Setup only enabled so exiting now")
		osExit(0)
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

	log.Println("Starting integration server commands API server")
	err = iscommandsapi.StartCommandsAPIServer(log, 7980, restartIntegrationServer)

	if err != nil {
		log.Println("Failed to start isapi server " + err.Error())
	} else {
		log.Println("Integration API started")
	}

	log.Println("Starting trace API server")
	err = trace.StartServer(log, 7981)
	if err != nil {
		log.Println("Failed to start trace API server, you will not be able to retrieve trace through the ACE dashboard " + err.Error())
	} else {
		log.Println("Trace API server started")
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
