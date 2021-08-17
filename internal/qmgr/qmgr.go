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
package qmgr

import (
	"io/ioutil"
	"log"
	"os"
	"time"
	"os/exec"

	"github.com/ot4i/ace-docker/internal/command"
	"github.com/ot4i/ace-docker/internal/logger"
)

// UseQueueManager returns a boolean for whether or not the system is using a queue manager.
func UseQueueManager() bool {
	useQmgrFlag, ok := os.LookupEnv("USE_QMGR")
	return ok && useQmgrFlag == "true"
}

// DevManager returns a boolean for whether or not to use developer edition MQ.
func DevManager() bool {
	devQmgrFlag, ok := os.LookupEnv("DEV_QMGR")
	return ok && devQmgrFlag == "true"
}

// StartQueueManager launches the runmqserver process in the background as the user "root".
// This returns a BackgroundCmd, wrapping the backgrounded process.
func StartQueueManager(log *logger.Logger) command.BackgroundCmd {
	if DevManager() {
		return command.RunBackground("runmqdevserver", log)
	} else {
		return command.RunBackground("runmqserver", log)
	}

}

// WaitForQueueManager will run the "chkmqready" command every 2 seconds until it returns
// an RC of zero, to indicate that the queue manager is ready.
func WaitForQueueManager(log *logger.Logger) error {
	for {
		//_, rc, err := command.RunAsUser("mqm", "chkmqready")

		//Fix for MQ 9.2
		cmd := exec.Command("chkmqready")
		_, rc, err := command.RunCmd(cmd)

		if rc != 0 || err != nil {
			log.Print("Queue manager not ready yet")
		}
		if rc == 0 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

// StopQueueManager will send a SIGINT to the runmqserver process, to signal the queue manager
// to stop, and then wait until the runmqserver process has ended.
func StopQueueManager(qmgrProcess command.BackgroundCmd) {
	if qmgrProcess.Cmd != nil && qmgrProcess.Started && !qmgrProcess.Finished {
		command.SigIntBackground(qmgrProcess)
		command.WaitOnBackground(qmgrProcess)
	}
}

// InitializeMQ will copy the mqsc scripts to the appropriate directory if supplied
func InitializeMQ() error {

	originalFile := "/home/aceuser/initial-config/mqsc/config.mqsc"
	if _, err := os.Stat(originalFile); err == nil {
		log.Println("Copying mqsc file ", originalFile)
		input, err := ioutil.ReadFile(originalFile)
		if err != nil {
			log.Println(err)
			return err
		}

		destinationFile := "/etc/mqm/config.mqsc"
		err = ioutil.WriteFile(destinationFile, input, 0644)
		if err != nil {
			log.Println("Error creating ", destinationFile)
			log.Println(err)
			return err
		}
		log.Println("Finished copying mqsc file")
	}

	log.Println("MQ initialization complete")
	return nil
}
