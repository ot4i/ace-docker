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

// chkacelively checks that ACE is still runing, by checking if the admin REST endpoint port is available.
// Additionally, if MQ is enabled, it also runs the MQ liveness probe chkmqhealthy.
package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ot4i/ace-docker/internal/command"
	"github.com/ot4i/ace-docker/internal/qmgr"
)

const restartIsTimeoutInSeconds = 60

func main() {
	// Check if the integration server has started the admin REST endpoint
	conn, err := net.Dial("tcp", "127.0.0.1:7600")

	if err != nil {

		fmt.Println("Unable to connect to IntegrationServer REST endpoint: " + err.Error() + ", ")

		fileInfo, statErr := os.Stat("/tmp/integration_server_restart.timestamp")

		if os.IsNotExist(statErr) {
			fmt.Println("Integration server is not active")
			os.Exit(1)
		} else if statErr != nil  {
			fmt.Println(statErr)
			os.Exit(1)
		} else {
			fmt.Println("Integration server restart file found")
			timeNow := time.Now()
			timeDiff := timeNow.Sub(fileInfo.ModTime())

			if timeDiff.Seconds() < restartIsTimeoutInSeconds {
				fmt.Println("Integration server is restarting")
				os.Exit(0)
			} else {
				fmt.Println("Integration restart time elapsed")
				os.Exit(1)
			}
		}
	}
	conn.Close()

	// Run MQ's liveness check chkmqhealthy, if enabled
	if qmgr.UseQueueManager() {
		out, rc, err := command.RunAsUser("mqm", "chkmqhealthy")
		if rc != 0 || err != nil {
			fmt.Println(out)
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
