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
package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"time"

	"github.com/ot4i/ace-docker/internal/command"
	"github.com/ot4i/ace-docker/internal/name"
	"github.com/ot4i/ace-docker/internal/qmgr"
	"gopkg.in/yaml.v2"
)

// createSystemQueues creates the default MQ service queues used by the Integration Server
func createSystemQueues() error {
	log.Println("Creating system queues")
	name, err := name.GetQueueManagerName()
	if err != nil {
		log.Errorf("Error getting queue manager name: %v", err)
		return err
	}

	out, _, err := command.Run("bash", "/opt/ibm/ace-11/server/sample/wmq/iib_createqueues.sh", name, "mqbrkrs")
	if err != nil {
		log.Errorf("Error creating system queues: %v", string(out))
		return err
	}
	log.Println("Created system queues")
	return nil
}

// initialIntegrationServerConfig walks through the /home/aceuser/initial-config directory
// looking for directories (each containing some config data), then runs a shell script
// called ace_config_{directory-name}.sh to process that data.
func initialIntegrationServerConfig() error {
	log.Printf("Performing initial configuration of integration server")

	getConfError := getConfigurationFromContentServer()
	if getConfError != nil {
		log.Errorf("Error getting configuration from content server: %v", getConfError)
		return getConfError
	}

	fileList, err := ioutil.ReadDir("/home/aceuser")
	if err != nil {
		log.Errorf("Error checking for an initial configuration folder: %v", err)
		return err
	}

	configDirExists := false
	for _, file := range fileList {
		if file.IsDir() && file.Name() == "initial-config" {
			configDirExists = true
		}
	}

	if !configDirExists {
		log.Printf("No initial configuration of integration server to perform")
		return nil
	}

	fileList, err = ioutil.ReadDir("/home/aceuser/initial-config")
	if err != nil {
		log.Errorf("Error checking for initial configuration folders: %v", err)
		return err
	}

	for _, file := range fileList {
		if file.IsDir() && file.Name() != "mqsc" {
			log.Printf("Processing configuration in folder %v", file.Name())

			//if qmgr.UseQueueManager() {
			//	out, _, err := command.RunAsUser("mqm", "ace_config_"+file.Name()+".sh")
			//	if err != nil {
			//		log.LogDirect(out)
			//		log.Errorf("Error processing configuration in folder %v: %v", file.Name(), err)
			//		return err
			//	}
			//	log.LogDirect(out)
			//} else {
			//	cmd := exec.Command("ace_config_"+file.Name()+".sh")
			//	out, _, err := command.RunCmd(cmd)
			//	if err != nil {
			//		log.LogDirect(out)
			//		log.Errorf("Error processing configuration in folder %v: %v", file.Name(), err)
			//		return err
			//	}
			//	log.LogDirect(out)
			//}

      		//Fix for MQ 9.2
			cmd := exec.Command("ace_config_"+file.Name()+".sh")
			out, _, err := command.RunCmd(cmd)

			if err != nil {
				log.LogDirect(out)
				log.Errorf("Error processing configuration in folder %v: %v", file.Name(), err)
				return err
			}

			log.LogDirect(out)
		}
	}

	enableMetrics := os.Getenv("ACE_ENABLE_METRICS")
	if enableMetrics == "true" || enableMetrics == "1" {
		enableMetricsError := enableMetricsInServerConf()
		if enableMetricsError != nil {
			log.Errorf("Error enabling metrics in server.conf.yaml: %v", enableMetricsError)
			return enableMetricsError
		}
	}

	enableOpenTracing := os.Getenv("ACE_ENABLE_OPEN_TRACING")
	if enableOpenTracing == "true" || enableOpenTracing == "1" {
		enableOpenTracingError := enableOpenTracingInServerConf()
		if enableOpenTracingError != nil {
			log.Errorf("Error enabling user exits in server.conf.yaml: %v", enableOpenTracingError)
			return enableOpenTracingError
		}
	}

	enableAdminssl := os.Getenv("ACE_ADMIN_SERVER_SECURITY")
	if enableAdminssl == "true" || enableAdminssl == "1" {
		enableAdminsslError := enableAdminsslInServerConf()
		if enableAdminsslError != nil {
			log.Errorf("Error enabling admin server security in server.conf.yaml: %v", enableAdminsslError)
			return enableAdminsslError
		}
	}

	log.Printf("Initial configuration of integration server complete")

	log.Println("Discovering override ports")

	out, _, err := command.Run("bash", "ace_discover_port_overrides.sh")
	if err != nil {
		log.Errorf("Error discovering override ports: %v", string(out))
		return err
	}
	log.Println("Successfully discovered override ports")

	return nil
}

// enableMetricsInServerConf adds Statistics fields to the server.conf.yaml in overrides
// If the file does not exist already it gets created.
func enableMetricsInServerConf() error {

	log.Println("Enabling metrics in server.conf.yaml")

	serverconfContent, readError := readServerConfFile()
	if readError != nil {
		if !os.IsNotExist(readError) {
			// Error is different from file not existing (if the file does not exist we will create it ourselves)
			log.Errorf("Error reading server.conf.yaml: %v", readError)
			return readError
		}
	}

	serverconfYaml, manipulationError := addMetricsToServerConf(serverconfContent)
	if manipulationError != nil {
		return manipulationError
	}

	writeError := writeServerConfFile(serverconfYaml)
	if writeError != nil {
		return writeError
	}

	log.Println("Metrics enabled in server.conf.yaml")

	return nil
}

// enableOpenTracingInServerConf adds OpenTracing UserExits fields to the server.conf.yaml in overrides
// If the file does not exist already it gets created.
func enableOpenTracingInServerConf() error {

	log.Println("Enabling OpenTracing in server.conf.yaml")

	serverconfContent, readError := readServerConfFile()
	if readError != nil {
		if !os.IsNotExist(readError) {
			// Error is different from file not existing (if the file does not exist we will create it ourselves)
			log.Errorf("Error reading server.conf.yaml: %v", readError)
			return readError
		}
	}

	serverconfYaml, manipulationError := addOpenTracingToServerConf(serverconfContent)
	if manipulationError != nil {
		return manipulationError
	}

	writeError := writeServerConfFile(serverconfYaml)
	if writeError != nil {
		return writeError
	}

	log.Println("OpenTracing enabled in server.conf.yaml")

	return nil
}

// enableAdminsslInServerConf adds RestAdminListener configuration fields to the server.conf.yaml in overrides
// based on the env vars ACE_ADMIN_SERVER_KEY, ACE_ADMIN_SERVER_CERT, ACE_ADMIN_SERVER_CA
// If the file does not exist already it gets created.
func enableAdminsslInServerConf() error {

	log.Println("Enabling Admin Server Security in server.conf.yaml")

	serverconfContent, readError := readServerConfFile()
	if readError != nil {
		if !os.IsNotExist(readError) {
			// Error is different from file not existing (if the file does not exist we will create it ourselves)
			log.Errorf("Error reading server.conf.yaml: %v", readError)
			return readError
		}
	}

	serverconfYaml, manipulationError := addAdminsslToServerConf(serverconfContent)
	if manipulationError != nil {
		return manipulationError
	}

	writeError := writeServerConfFile(serverconfYaml)
	if writeError != nil {
		return writeError
	}

	log.Println("Admin Server Security enabled in server.conf.yaml")

	return nil
}

// readServerConfFile returns the content of the server.conf.yaml file in the overrides folder
func readServerConfFile() ([]byte, error) {
	content, err := ioutil.ReadFile("/home/aceuser/ace-server/overrides/server.conf.yaml")
	return content, err
}

// writeServerConfFile writes the yaml content to the server.conf.yaml file in the overrides folder
// It creates the file if it doesn't already exist
func writeServerConfFile(content []byte) error {
	writeError := ioutil.WriteFile("/home/aceuser/ace-server/overrides/server.conf.yaml", content, 0644)
	if writeError != nil {
		log.Errorf("Error writing server.conf.yaml: %v", writeError)
		return writeError
	}
	return nil
}

// addMetricsToServerConf gets the content of the server.conf.yaml and adds the metrics fields to it
// It returns the updated server.conf.yaml content
func addMetricsToServerConf(serverconfContent []byte) ([]byte, error) {
	serverconfMap := make(map[interface{}]interface{})
	unmarshallError := yaml.Unmarshal([]byte(serverconfContent), &serverconfMap)
	if unmarshallError != nil {
		log.Errorf("Error unmarshalling server.conf.yaml: %v", unmarshallError)
		return nil, unmarshallError
	}

	snapshotObj := map[string]string{
		"publicationOn":    "active",
		"nodeDataLevel":    "basic",
		"outputFormat":     "json",
		"threadDataLevel":  "none",
		"accountingOrigin": "none",
	}

	resourceObj := map[string]bool{
		"reportingOn": true,
	}

	if serverconfMap["Statistics"] != nil {
		statistics := serverconfMap["Statistics"].(map[interface{}]interface{})

		if statistics["Snapshot"] != nil {
			snapshot := statistics["Snapshot"].(map[interface{}]interface{})
			snapshot["publicationOn"] = "active"
			snapshot["nodeDataLevel"] = "basic"
			snapshot["outputFormat"] = "json"
			snapshot["threadDataLevel"] = "none"
		} else {
			statistics["Snapshot"] = snapshotObj
		}

		statistics["Resource"] = resourceObj

	} else {
		serverconfMap["Statistics"] = map[string]interface{}{
			"Snapshot": snapshotObj,
			"Resource": resourceObj,
		}
	}

	serverconfYaml, marshallError := yaml.Marshal(&serverconfMap)
	if marshallError != nil {
		log.Errorf("Error marshalling server.conf.yaml: %v", marshallError)
		return nil, marshallError
	}

	return serverconfYaml, nil
}

// addOpenTracingToServerConf gets the content of the server.conf.yaml and adds the OpenTracing UserExits fields to it
// It returns the updated server.conf.yaml content
func addOpenTracingToServerConf(serverconfContent []byte) ([]byte, error) {
	serverconfMap := make(map[interface{}]interface{})
	unmarshallError := yaml.Unmarshal([]byte(serverconfContent), &serverconfMap)
	if unmarshallError != nil {
		log.Errorf("Error unmarshalling server.conf.yaml: %v", unmarshallError)
		return nil, unmarshallError
	}

	if serverconfMap["UserExits"] != nil {
		userExits := serverconfMap["UserExits"].(map[string]string)

		userExits["activeUserExitList"] = "ACEOpenTracingUserExit"
		userExits["userExitPath"] = "/opt/ACEOpenTracing"

	} else {
		serverconfMap["UserExits"] = map[string]string{
			"activeUserExitList": "ACEOpenTracingUserExit",
			"userExitPath":       "/opt/ACEOpenTracing",
		}
	}

	serverconfYaml, marshallError := yaml.Marshal(&serverconfMap)
	if marshallError != nil {
		log.Errorf("Error marshalling server.conf.yaml: %v", marshallError)
		return nil, marshallError
	}

	return serverconfYaml, nil
}

// addAdminsslToServerConf gets the content of the server.conf.yaml and adds the Admin Server Security fields to it
// It returns the updated server.conf.yaml content
func addAdminsslToServerConf(serverconfContent []byte) ([]byte, error) {
	serverconfMap := make(map[interface{}]interface{})
	unmarshallError := yaml.Unmarshal([]byte(serverconfContent), &serverconfMap)
	if unmarshallError != nil {
		log.Errorf("Error unmarshalling server.conf.yaml: %v", unmarshallError)
		return nil, unmarshallError
	}

	// Get the keys, certs location and default if not found
	cert := os.Getenv("ACE_ADMIN_SERVER_CERT")
	if cert == "" {
		cert = "/home/aceuser/adminssl/tls.crt.pem"
	}

	key := os.Getenv("ACE_ADMIN_SERVER_KEY")
	if key == "" {
		key = "/home/aceuser/adminssl/tls.key.pem"
	}

	cacert := os.Getenv("ACE_ADMIN_SERVER_CA")
	if cacert == "" {
		cacert = "/home/aceuser/adminssl"
	}

	isTrue := true
	// Only update if there is not an existing entry in the override server.conf.yaml
	// so we don't overwrite any customer provided configuration
	if serverconfMap["RestAdminListener"] == nil {
		serverconfMap["RestAdminListener"] = map[string]interface{}{
			"sslCertificate" : cert,
			"sslPassword" : key,
			"requireClientCert" : isTrue,
			"caPath" : cacert,
		}
		log.Printf("Admin Server Security updating RestAdminListener using ACE_ADMIN_SERVER environment variables")
	} else {
		restAdminListener := serverconfMap["RestAdminListener"].(map[interface{}]interface{})

    	if restAdminListener["sslCertificate"] == nil {
    		restAdminListener["sslCertificate"] =  cert
    	}
    	if restAdminListener["sslPassword"] == nil {
    		restAdminListener["sslPassword"] =  key
    	}
    	if restAdminListener["requireClientCert"] == nil {
    		restAdminListener["requireClientCert"] = isTrue
    	}
    	if restAdminListener["caPath"] == nil {
    		restAdminListener["caPath"] = cacert
    	}
		log.Printf("Admin Server Security merging RestAdminListener using ACE_ADMIN_SERVER environment variables")
	}

	serverconfYaml, marshallError := yaml.Marshal(&serverconfMap)
	if marshallError != nil {
		log.Errorf("Error marshalling server.conf.yaml: %v", marshallError)
		return nil, marshallError
	}

	return serverconfYaml, nil
}

// getConfigurationFromContentServer checks if ACE_CONTENT_SERVER_URL exists.  If so then it pulls
// a bar file from that URL
func getConfigurationFromContentServer() error {

	url := os.Getenv("ACE_CONTENT_SERVER_URL")
	if url == "" {
		log.Printf("No content server url available")
		return nil
	}

	defaultContentServer := os.Getenv("DEFAULT_CONTENT_SERVER")
	if defaultContentServer == "" {
		log.Printf("Can't tell if content server is default one so defaulting")
		defaultContentServer = "true"
	}

	serverName := os.Getenv("ACE_CONTENT_SERVER_NAME")
	if serverName == "" {
		log.Printf("No content server name available but a url is defined")
		return errors.New("No content server name available but a url is defined")
	}

	token := os.Getenv("ACE_CONTENT_SERVER_TOKEN")
	if token == "" && defaultContentServer == "true" {
		log.Errorf("No content server token available but a url is defined")
		return errors.New("No content server token available but a url is defined")
	}

	err := os.Mkdir("/home/aceuser/initial-config/bars", os.ModePerm)
	if err != nil {
		log.Errorf("Error creating directory /home/aceuser/initial-config/bars: %v", err)
		return err
	}

	file, err := os.Create("/home/aceuser/initial-config/bars/barfile.bar")
	if err != nil {
		log.Errorf("Error creating file %v: %v", file, err)
		return err
	}
	defer file.Close()

	// Create a CA certificate pool and add cacert to it
	var contentServerCACert string
	if defaultContentServer == "true" {
		log.Printf("Getting configuration from content server")
		contentServerCACert = "/home/aceuser/ssl/cacert.pem"
		url = url + "?archive=true"
	} else {
		log.Printf("Getting configuration from custom content server")
		contentServerCACert = os.Getenv("CONTENT_SERVER_CA")
		if contentServerCACert == "" {
			log.Printf("CONTENT_SERVER_CA not defined")
			return errors.New("CONTENT_SERVER_CA not defined")
		}
	}
	log.Printf("Using ca file %s", contentServerCACert)
	caCert, err := ioutil.ReadFile(contentServerCACert)
	if err != nil {
		log.Errorf("Error reading CA Certificate")
		return errors.New("Error reading CA Certificate")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// If provided read the key pair to create certificate
	contentServerCert := os.Getenv("CONTENT_SERVER_CERT")
	contentServerKey := os.Getenv("CONTENT_SERVER_KEY")
	cert, err := tls.LoadX509KeyPair(contentServerCert, contentServerKey)
	if err != nil {
		if contentServerCert != "" && contentServerKey != "" {
			log.Errorf("Error reading Certificates: %s", err)
			return errors.New("Error reading Certificates")
		}
	} else {
		log.Printf("Using certs for mutual auth")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
				ServerName:   serverName,
			},
		},
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error creating request for content server")
		return err
	}

	request.Header.Set("x-ibm-ace-directory-token", token)
	response, err := client.Do(request)
	if err != nil {
		log.Errorf("Error downloading from %v: %v", url, err)
		return err
	}
	if response.StatusCode != 200 {
		log.Errorf("Error downloading from %v: %v", url, response.Status)
		return err
	}

	defer response.Body.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Errorf("Error writing file %v: %v", file, err)
		return err
	}

	log.Printf("Configuration pulled from content server successfully")
	return nil
}

// startIntegrationServer launches the IntegrationServer process in the background as the user "aceuser".
// This returns a BackgroundCmd, wrapping the backgrounded process, or an error if we completely failed to
// start the process
func startIntegrationServer() command.BackgroundCmd {
	logOutputFormat := getLogOutputFormat()

	serverName, err := name.GetIntegrationServerName()
	if err != nil {
		log.Printf("Error getting integration server name: %v", err)
		returnErr := command.BackgroundCmd{}
		returnErr.ReturnCode = -1
		returnErr.ReturnError = err
		return returnErr
	}

	defaultAppName := os.Getenv("ACE_DEFAULT_APPLICATION_NAME")
	if defaultAppName == "" {
		log.Printf("No default application name supplied. Using the integration server name instead.")
		defaultAppName = serverName
	}

	if qmgr.UseQueueManager() {
		qmgrName, err := name.GetQueueManagerName()
		if err != nil {
			log.Printf("Error getting queue manager name: %v", err)
			returnErr := command.BackgroundCmd{}
			returnErr.ReturnCode = -1
			returnErr.ReturnError = err
			return returnErr
		}

		//return command.RunAsUserBackground("mqm", "ace_integration_server.sh", log, "-w", "/home/aceuser/ace-server", "--name", serverName, "--mq-queue-manager-name", qmgrName, "--log-output-format", logOutputFormat, "--console-log", "--default-application-name", defaultAppName)

    	//Fix for MQ 9.2
		return command.RunBackground("ace_integration_server.sh", log, "-w", "/home/aceuser/ace-server", "--name", serverName, "--mq-queue-manager-name", qmgrName, "--log-output-format", logOutputFormat, "--console-log", "--default-application-name", defaultAppName)
	}

	thisUser, err := user.Current()
	if err != nil {
		log.Errorf("Error finding this user: %v", err)
		returnErr := command.BackgroundCmd{}
		returnErr.ReturnCode = -1
		returnErr.ReturnError = err
		return returnErr
	}

	return command.RunAsUserBackground(thisUser.Username, "ace_integration_server.sh", log, "-w", "/home/aceuser/ace-server", "--name", serverName, "--log-output-format", logOutputFormat, "--console-log", "--default-application-name", defaultAppName)
}

func waitForIntegrationServer() error {
	for {

		//if qmgr.UseQueueManager() {
		//	_, rc, err := command.RunAsUser("mqm", "chkaceready")
		//	if rc != 0 || err != nil {
		//		log.Printf("Integration server not ready yet")
		//	}
		//	if rc == 0 {
		//		break
		//	}
		//	time.Sleep(5 * time.Second)
		//} else {
		//	cmd := exec.Command("chkaceready")
		//	_, rc, err := command.RunCmd(cmd)
		//	if rc != 0 || err != nil {
		//		log.Printf("Integration server not ready yet")
		//	}
		//	if rc == 0 {
		//		break
		//	}
		//	time.Sleep(5 * time.Second)
		//}

    	//Fix for MQ 9.2
		cmd := exec.Command("chkaceready")
		_, rc, err := command.RunCmd(cmd)

		if rc != 0 || err != nil {
			log.Printf("Integration server not ready yet")
		}

		if rc == 0 {
			break
		}

		time.Sleep(5 * time.Second)
	}
	return nil
}

func stopIntegrationServer(integrationServerProcess command.BackgroundCmd) {
	if integrationServerProcess.Cmd != nil && integrationServerProcess.Started && !integrationServerProcess.Finished {
		command.SigIntBackground(integrationServerProcess)
		command.WaitOnBackground(integrationServerProcess)
	}
}

func createWorkDir() error {
	log.Printf("Checking if work dir is already initialized")
	f, err := os.Open("/home/aceuser/ace-server")
	if err != nil {
		log.Printf("Error reading /home/aceuser/ace-server")
		return err
	}

	log.Printf("Checking for contents in the work dir")
	_, err = f.Readdirnames(1)
	if err != nil {
		log.Printf("Work dir is not yet initialized - initializing now in /home/aceuser/ace-server")

		if qmgr.UseQueueManager() {
			_, _, err := command.RunAsUser("1000", "/opt/ibm/ace-11/server/bin/mqsicreateworkdir", "/home/aceuser/ace-server")
			if err != nil {
				log.Printf("Error reading initializing work dir")
				return err
			}
		} else {
			cmd := exec.Command("/opt/ibm/ace-11/server/bin/mqsicreateworkdir", "/home/aceuser/ace-server")
			_, _, err := command.RunCmd(cmd)
			if err != nil {
				log.Printf("Error reading initializing work dir")
				return err
			}
		}
	}
	log.Printf("Work dir initialization complete")
	return nil
}

func checkLogs() error {
	log.Printf("Contents of log directory")
	system("ls", "-l", "/home/aceuser/ace-server/config/common/log")

	if os.Getenv("MQSI_PREVENT_CONTAINER_SHUTDOWN") == "true" {
		log.Printf("MQSI_PREVENT_CONTAINER_SHUTDOWN set to blocking container shutdown to enable log copy out")
		log.Printf("Once all logs have been copied out please kill container")
		select {}
	}

	log.Printf("If you want to stop the container shutting down to enable retrieval of these files please set the environment variable \"MQSI_PREVENT_CONTAINER_SHUTDOWN=true\"")
	log.Printf("If you are running under kubernetes you will also need to disable the livenessProbe")
	log.Printf("Log checking complete")
	return nil
}

func system(cmd string, arg ...string) {
	out, err := exec.Command(cmd, arg...).Output()
	if err != nil {
		log.Printf(err.Error())
	}
	log.Printf(string(out))
}
