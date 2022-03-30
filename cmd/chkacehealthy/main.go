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
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
)

var checkACE = checkACElocal
var httpCheck = httpChecklocal
var socketCheck = socketChecklocal
var osExit = os.Exit

const restartIsTimeoutInSeconds = 60

var netDial = net.Dial
var httpGet = http.Get

func main() {

	err := checkACE()
	if err != nil {
		log.Fatal(err)
	}

	// If knative service also check FDR is up
	knative := os.Getenv("KNATIVESERVICE")
	if knative == "true" || knative == "1" {
		fmt.Println("KNATIVESERVICE set so checking FDR container")
		err := checkDesignerHealth()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("KNATIVESERVICE is not set so skipping FDR checks")
	}

}

func checkDesignerHealth() error {
	// HTTP LMAP endpoint
	err := httpCheck("LMAP Port", "http://localhost:3002/admin/ready")
	if err != nil {
		return err
	}

	isConnectorService := os.Getenv("CONNECTOR_SERVICE")
	if isConnectorService == "true" || isConnectorService == "1" {
		// HTTP LCP Connector service endpoint
		err = httpCheck("LCP Port", "http://localhost:3001/admin/ready")
		if err != nil {
			return err
		}
	}

	// LCP api flow endpoint
	lcpsocket := "/tmp/lcp.socket"
	if value, ok := os.LookupEnv("LCP_IPC_PATH"); ok {
		lcpsocket = value
	}
	err = socketCheck("LCP socket", lcpsocket)
	if err != nil {
		return err
	}

	// LMAP endpoint
	lmapsocket := "/tmp/lmap.socket"
	if value, ok := os.LookupEnv("LMAP_IPC_PATH"); ok {
		lmapsocket = value
	}
	err = socketCheck("LMAP socket", lmapsocket)
	if err != nil {
		return err
	}

	return nil
}

func isEnvExist(key string) bool {
	if _, ok := os.LookupEnv(key); ok {
		return true
	}
	return false
}

func checkACElocal() error {
	// Check if the integration server has started the admin REST endpoint
	conn, err := netDial("tcp", "127.0.0.1:7600")

	if err != nil {

		fmt.Println("Unable to connect to IntegrationServer REST endpoint: " + err.Error() + ", ")

		fileInfo, statErr := os.Stat("/tmp/integration_server_restart.timestamp")

		if os.IsNotExist(statErr) {
			fmt.Println("Integration server is not active")
			return errors.NewBadRequest("Integration server is not active")
		} else if statErr != nil {
			fmt.Println(statErr)
			return errors.NewBadRequest("stat error " + statErr.Error())
		} else {
			fmt.Println("Integration server restart file found")
			timeNow := time.Now()
			timeDiff := timeNow.Sub(fileInfo.ModTime())

			if timeDiff.Seconds() < restartIsTimeoutInSeconds {
				fmt.Println("Integration server is restarting")
			} else {
				fmt.Println("Integration restart time elapsed")
				return errors.NewBadRequest("Integration restart time elapsed")
			}
		}
	} else {
		fmt.Println("ACE ready check passed")
	}
	conn.Close()
	return nil
}

func httpChecklocal(name string, addr string) error {
	resp, err := httpGet(addr)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		fmt.Println(name + " ready check failed - HTTP Status is not 200 range")
		return errors.NewBadRequest(name + " ready check failed - HTTP Status is not 200 range")
	} else {
		fmt.Println(name + " ready check passed")
	}
	return nil
}

func socketChecklocal(name string, socket string) error {
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socket)
			},
		},
	}
	response, err := httpc.Get("http://dummyHostname/admin/ready")
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		log.Fatal(name + " ready check failed - HTTP Status is not 200 range")
		return errors.NewBadRequest(name + " ready check failed - HTTP Status is not 200 range")
	} else {
		fmt.Println(name + " ready check passed")
	}
	return nil
}
