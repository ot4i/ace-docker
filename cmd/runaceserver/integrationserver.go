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
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ot4i/ace-docker/common/contentserver"
	"github.com/ot4i/ace-docker/internal/command"
	"github.com/ot4i/ace-docker/internal/configuration"
	"github.com/ot4i/ace-docker/internal/name"
	"github.com/ot4i/ace-docker/internal/qmgr"
	"github.com/ot4i/ace-docker/internal/webadmin"
	"gopkg.in/yaml.v2"

	"software.sslmate.com/src/go-pkcs12"
)

var osMkdir = os.Mkdir
var osCreate = os.Create
var osStat = os.Stat
var ioutilReadFile = ioutil.ReadFile
var ioutilReadDir = ioutil.ReadDir
var ioCopy = io.Copy
var contentserverGetBAR = contentserver.GetBAR
var watcher *fsnotify.Watcher
var createSHAServerConfYaml = createSHAServerConfYamlLocal
var homedir string = "/home/aceuser"
var initialConfigDir string = "/home/aceuser/initial-config"
var ConfigureWebAdminUsers = webadmin.ConfigureWebAdminUsers
var readServerConfFile = readServerConfFileLocal
var yamlUnmarshal = yaml.Unmarshal
var yamlMarshal = yaml.Marshal
var writeServerConfFile = writeServerConfFileLocal
var getConfigurationFromContentServer = getConfigurationFromContentServerLocal

// createSystemQueues creates the default MQ service queues used by the Integration Server
func createSystemQueues() error {
	log.Println("Creating system queues")
	name, err := name.GetQueueManagerName()
	if err != nil {
		log.Errorf("Error getting queue manager name: %v", err)
		return err
	}

	out, _, err := command.Run("bash", "/opt/ibm/ace-12/server/sample/wmq/iib_createqueues.sh", name, "mqbrkrs")
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

	if configuration.ContentServer {
		getConfError := getConfigurationFromContentServer()
		if getConfError != nil {
			log.Errorf("Error getting configuration from content server: %v", getConfError)
			return getConfError
		}
	}

	fileList, err := ioutilReadDir(homedir)
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

	fileList, err = ioutil.ReadDir(initialConfigDir)
	if err != nil {
		log.Errorf("Error checking for initial configuration folders: %v", err)
		return err
	}

	// Sort filelist to server.conf.yaml gets written before webusers are processedconfigDirExists
	SortFileNameAscend(fileList)
	for _, file := range fileList {
		if file.IsDir() && file.Name() != "mqsc" && file.Name() != "workdir_overrides" {
			log.Printf("Processing configuration in folder %v", file.Name())
			if qmgr.UseQueueManager() {
				out, _, err := command.RunAsUser("mqm", "ace_config_"+file.Name()+".sh")
				if err != nil {
					log.LogDirect(out)
					log.Errorf("Error processing configuration in folder %v: %v", file.Name(), err)
					return err
				}
				log.LogDirect(out)
			} else {
				if file.Name() == "webusers" {
					log.Println("Configuring server.conf.yaml overrides - Webadmin")
					updateServerConf := createSHAServerConfYaml()
					if updateServerConf != nil {
						log.Errorf("Error setting webadmin SHA server.conf.yaml: %v", updateServerConf)
						return updateServerConf
					}
					log.Println("Configuring WebAdmin Users")
					err := ConfigureWebAdminUsers(log)
					if err != nil {
						log.Errorf("Error configuring the WebAdmin users : %v", err)
						return err
					}
				}
				if file.Name() != "webusers" {
					cmd := exec.Command("ace_config_" + file.Name() + ".sh")
					out, _, err := command.RunCmd(cmd)
					if err != nil {
						log.LogDirect(out)
						log.Errorf("Error processing configuration in folder %v: %v", file.Name(), err)
						return err
					}
					log.LogDirect(out)
				}
			}
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

	forceFlowHttps := os.Getenv("FORCE_FLOW_HTTPS")
	if forceFlowHttps == "true" || forceFlowHttps == "1" {
		log.Printf("Forcing all flows to be https. FORCE_FLOW_HTTPS=%v", forceFlowHttps)

		// create the https nodes keystore and password
		password := generatePassword(10)

		log.Println("Force Flows to be HTTPS running keystore creation commands")
		cmd := exec.Command("ace_forceflowhttps.sh", password)
		out, _, err := command.RunCmd(cmd)
		if err != nil {
			log.Errorf("Error creating force flow https keystore and password, retrying. Error is %v", err)
			log.LogDirect(out)
			return err
		}

		log.Println("Force Flows to be HTTPS in server.conf.yaml")
		forceFlowHttpsError := forceFlowsHttpsInServerConf()
		if forceFlowHttpsError != nil {
			log.Errorf("Error forcing flows to https in server.conf.yaml: %v", forceFlowHttpsError)
			return forceFlowHttpsError
		}

		// Start watching the ..data/tls.key file where the tls.key secret is stored and if it changes recreate the p12 keystore and restart the HTTPSConnector dynamically
		// need to watch the mounted ..data/tls.key file here as the tls.key symlink timestamp never changes
		log.Println("Force Flows to be HTTPS starting to watch /home/aceuser/httpsNodeCerts/..data/tls.key")
		watcher = watchForceFlowsHTTPSSecret(password)
		err = watcher.Add("/home/aceuser/httpsNodeCerts/..data/tls.key")
		if err != nil {
			log.Errorf("Error watching /home/aceuser/httpsNodeCerts/tls.key for Force Flows to be HTTPS: %v", err)
		}
	} else {
		log.Printf("Not Forcing all flows to be https as FORCE_FLOW_HTTPS=%v", forceFlowHttps)
	}

	log.Println("Initial configuration of integration server complete")

	return nil
}

func SortFileNameAscend(files []os.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
}

func createSHAServerConfYamlLocal() error {

	oldserverconfContent, readError := readServerConfFile()
	if readError != nil {
		if !os.IsNotExist(readError) {
			// Error is different from file not existing (if the file does not exist we will create it ourselves)
			log.Errorf("Error reading server.conf.yaml: %v", readError)
			return readError
		}
	}

	serverconfMap := make(map[interface{}]interface{})
	unmarshallError := yamlUnmarshal([]byte(oldserverconfContent), &serverconfMap)
	if unmarshallError != nil {
		log.Errorf("Error unmarshalling server.conf.yaml: %v", unmarshallError)
		return unmarshallError
	}

	if serverconfMap["RestAdminListener"] == nil {
		serverconfMap["RestAdminListener"] = map[string]interface{}{
			"authorizationEnabled": true,
			"authorizationMode":    "file",
			"basicAuth":            true,
		}
	} else {
		restAdminListener := serverconfMap["RestAdminListener"].(map[interface{}]interface{})
		restAdminListener["authorizationEnabled"] = true
		restAdminListener["authorizationMode"] = "file"
		restAdminListener["basicAuth"] = true

	}

	serverconfYaml, marshallError := yamlMarshal(&serverconfMap)
	if marshallError != nil {
		log.Errorf("Error marshalling server.conf.yaml: %v", marshallError)
		return marshallError
	}
	writeError := writeServerConfFile(serverconfYaml)
	if writeError != nil {
		return writeError
	}

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

// readServerConfFile returns the content of the server.conf.yaml file in the overrides folder
func readServerConfFileLocal() ([]byte, error) {
	content, err := ioutil.ReadFile("/home/aceuser/ace-server/overrides/server.conf.yaml")
	return content, err
}

// writeServerConfFile writes the yaml content to the server.conf.yaml file in the overrides folder
// It creates the file if it doesn't already exist
func writeServerConfFileLocal(content []byte) error {
	writeError := ioutil.WriteFile("/home/aceuser/ace-server/overrides/server.conf.yaml", content, 0644)
	if writeError != nil {
		log.Errorf("Error writing server.conf.yaml: %v", writeError)
		return writeError
	}
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

// addMetricsToServerConf gets the content of the server.conf.yaml and adds the metrics fields to it
// It returns the updated server.conf.yaml content
func addMetricsToServerConf(serverconfContent []byte) ([]byte, error) {
	serverconfMap := make(map[interface{}]interface{})
	unmarshallError := yamlUnmarshal([]byte(serverconfContent), &serverconfMap)
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
			if snapshot["publicationOn"] == nil {
				snapshot["publicationOn"] = "active"
			}
			if snapshot["nodeDataLevel"] == nil {
				snapshot["nodeDataLevel"] = "basic"
			}
			if snapshot["outputFormat"] == nil {
				snapshot["outputFormat"] = "json"
			} else {
				snapshot["outputFormat"] = fmt.Sprintf("%v", snapshot["outputFormat"]) + ",json"
			}
			if snapshot["threadDataLevel"] == nil {
				snapshot["threadDataLevel"] = "none"
			}
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
			"sslCertificate":    cert,
			"sslPassword":       key,
			"requireClientCert": isTrue,
			"caPath":            cacert,
		}
		log.Printf("Admin Server Security updating RestAdminListener using ACE_ADMIN_SERVER environment variables")
	} else {
		restAdminListener := serverconfMap["RestAdminListener"].(map[interface{}]interface{})

		if restAdminListener["sslCertificate"] == nil {
			restAdminListener["sslCertificate"] = cert
		}
		if restAdminListener["sslPassword"] == nil {
			restAdminListener["sslPassword"] = key
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
func getConfigurationFromContentServerLocal() error {

	// ACE_CONTENT_SERVER_URL can contain 1 or more comma separated urls
	urls := os.Getenv("ACE_CONTENT_SERVER_URL")
	if urls == "" {
		log.Printf("No content server url available")
		return nil
	}

	defaultContentServer := os.Getenv("DEFAULT_CONTENT_SERVER")
	if defaultContentServer == "" {
		log.Printf("Can't tell if content server is default one so defaulting")
		defaultContentServer = "true"
	}

	err := osMkdir("/home/aceuser/initial-config/bars", os.ModePerm)
	if err != nil {
		log.Errorf("Error creating directory /home/aceuser/initial-config/bars: %v", err)
		return err
	}

	// check for AUTH env parameters (needed if auth is not encoded in urls for backward compatibility pending operator changes)
	envServerName := os.Getenv("ACE_CONTENT_SERVER_NAME")
	envToken := os.Getenv("ACE_CONTENT_SERVER_TOKEN")

	urlArray := strings.Split(urls, ",")
	for _, barurl := range urlArray {

		serverName := envServerName
		token := envToken

		// the ace content server name is the name of the secret where this cert is
		// eg. secretName: {{ index (splitList ":" (index (splitList "/" (trim .Values.contentServerURL)) 2)) 0 | quote }} ?
		//  https://domsdash-ibm-ace-dashboard-prod:3443/v1/directories/CustomerDatabaseV1?userid=fsdjfhksdjfhsd
		//  or https://test-acecontentserver-ace-dom.svc:3443/v1/directories/testdir?e31d23f6-e3ba-467d-ab3b-ceb0ab12eead
		// Mutli-tenant : https://test-acecontentserver-ace-dom.svc:3443/v1/namespace/directories/testdir
		// https://dataplane-api-dash.appconnect:3443/v1/appc-fakeid/directories/ace_manualtest_callableflow

		splitOnSlash := strings.Split(barurl, "/")

		if len(splitOnSlash) > 2 {
			serverName = strings.Split(splitOnSlash[2], ":")[0] // test-acecontentserver.ace-dom
		} else {
			// if we have not found serverName from either env or url error
			log.Printf("No content server name available but a url is defined - Have you forgotten to define your BAR_AUTH configuration resource?")
			return errors.New("No content server name available but a url is defined  - Have you forgotten to define your BAR_AUTH configuration resource?")
		}

		// if ACE_CONTENT_SERVER_TOKEN was set use it. It may have been read from a secret
		// otherwise then look in the url for ?
		if token == "" {
			splitOnQuestion := strings.Split(barurl, "?")
			if len(splitOnQuestion) > 1 && splitOnQuestion[1] != "" {
				barurl = splitOnQuestion[0] // https://test-acecontentserver.ace-dom.svc:3443/v1/directories/testdir
				token = splitOnQuestion[1]  //userid=fsdjfhksdjfhsd
			} else if defaultContentServer == "true" {
				// if we have not found token from either env or url error
				log.Errorf("No content server token available but a url is defined")
				return errors.New("No content server token available but a url is defined")
			}
		}

		// use the last part of the url path (base) for the filename
		u, err := url.Parse(barurl)
		if err != nil {
			log.Errorf("Error parsing content server url : %v", err)
			return err
		}

		var filename string
		if len(urlArray) == 1 {
			// temporarily override the bar name  with "barfile.bar" if we only have ONE bar file until mq connector is fixed to support any bar name
			filename = "/home/aceuser/initial-config/bars/barfile.bar"
		} else {
			// Multiple bar support. Need to loop to check that the file does not already exist
			// (case where multiple bars have the same name)
			isAvailable := false
			count := 0
			for !isAvailable {
				if count == 0 {
					filename = "/home/aceuser/initial-config/bars/" + path.Base(u.Path) + ".bar"
				} else {
					filename = "/home/aceuser/initial-config/bars/" + path.Base(u.Path) + "-" + fmt.Sprint(count) + ".bar"
					log.Printf("Previous path already in use. Testing filename: " + filename)
				}

				if _, err := osStat(filename); os.IsNotExist(err) {
					log.Printf("No existing file on that path so continuing")
					isAvailable = true
				}
				count++
			}
		}

		log.Printf("Will save bar as: " + filename)

		file, err := osCreate(filename)
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
			barurl = barurl + "?archive=true"
		} else {
			log.Printf("Getting configuration from custom content server")
			contentServerCACert = os.Getenv("CONTENT_SERVER_CA")
			if token != "" {
				barurl = barurl + "?" + token
			}
			if contentServerCACert == "" {
				log.Printf("CONTENT_SERVER_CA not defined")
				return errors.New("CONTENT_SERVER_CA not defined")
			}
		}
		log.Printf("Using the following url: " + barurl)

		log.Printf("Using ca file %s", contentServerCACert)
		caCert, err := ioutilReadFile(contentServerCACert)
		if err != nil {
			log.Errorf("Error reading CA Certificate")
			return errors.New("Error reading CA Certificate")
		}

		contentServerCert := os.Getenv("CONTENT_SERVER_CERT")
		contentServerKey := os.Getenv("CONTENT_SERVER_KEY")

		bar, err := contentserverGetBAR(barurl, serverName, token, caCert, contentServerCert, contentServerKey, log)
		if err != nil {
			return err
		}
		defer bar.Close()

		_, err = ioCopy(file, bar)
		if err != nil {
			log.Errorf("Error writing file %v: %v", file, err)
			return err
		}

		log.Printf("Configuration pulled from content server successfully")

	}
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
		return command.RunAsUserBackground("mqm", "ace_integration_server.sh", log, "-w", "/home/aceuser/ace-server", "--name", serverName, "--mq-queue-manager-name", qmgrName, "--log-output-format", logOutputFormat, "--console-log", "--default-application-name", defaultAppName)
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
		if qmgr.UseQueueManager() {
			_, rc, err := command.RunAsUser("mqm", "chkaceready")
			if rc != 0 || err != nil {
				log.Printf("Integration server not ready yet")
			}
			if rc == 0 {
				break
			}
			time.Sleep(5 * time.Second)
		} else {
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
			_, _, err := command.RunAsUser("mqm", "/opt/ibm/ace-12/server/bin/mqsicreateworkdir", "/home/aceuser/ace-server")
			if err != nil {
				log.Printf("Error reading initializing work dir")
				return err
			}
		} else {
			cmd := exec.Command("/opt/ibm/ace-12/server/bin/mqsicreateworkdir", "/home/aceuser/ace-server")
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

// applyWorkdirOverrides walks through the home/aceuser/initial-config/workdir_overrides directory
// we want to do this here rather than the loop above as we want to make sure we have done everything
// else before applying the workdir overrides and then start the integration server
func applyWorkdirOverrides() error {

	fileList, err := ioutil.ReadDir("/home/aceuser")
	if err != nil {
		log.Errorf("Error checking for the aceuser home directoy: %v", err)
		return err
	}

	configDirExists := false
	for _, file := range fileList {
		if file.IsDir() && file.Name() == "initial-config" {
			configDirExists = true
		}
	}

	if !configDirExists {
		log.Printf("No initial-config directory found")
		return nil
	}

	fileList, err = ioutil.ReadDir("/home/aceuser/initial-config")
	if err != nil {
		log.Errorf("Error checking for initial configuration folders: %v", err)
		return err
	}

	for _, file := range fileList {
		if file.IsDir() && file.Name() == "workdir_overrides" {
			log.Println("Applying workdir overrides to the integration server")
			cmd := exec.Command("ace_config_workdir_overrides.sh")
			out, _, err := command.RunCmd(cmd)
			log.LogDirect(out)
			if err != nil {
				log.Errorf("Error processing workdir overrides in folder %v: %v", file.Name(), err)
				return err
			}
			log.Printf("Workdir overrides applied to the integration server complete")
		}
	}

	return nil
}

// forceFlowHttps adds ResourceManagers HTTPSConnector fields to the server.conf.yaml in overrides using the keystore and password created
// If the file does not exist already it gets created.
func forceFlowsHttpsInServerConf() error {
	serverconfContent, readError := readServerConfFile()
	if readError != nil {
		if !os.IsNotExist(readError) {
			// Error is different from file not existing (if the file does not exist we will create it ourselves)
			log.Errorf("Error reading server.conf.yaml: %v", readError)
			return readError
		}
	}

	serverconfYaml, manipulationError := addforceFlowsHttpsToServerConf(serverconfContent)
	if manipulationError != nil {
		return manipulationError
	}

	writeError := writeServerConfFile(serverconfYaml)
	if writeError != nil {
		return writeError
	}
	log.Println("Force Flows to be HTTPS in server.conf.yaml completed")

	return nil
}

// addforceFlowsHttpsToServerConf gets the content of the server.conf.yaml and adds the Force Flow Security fields to it
// It returns the updated server.conf.yaml content
func addforceFlowsHttpsToServerConf(serverconfContent []byte) ([]byte, error) {
	serverconfMap := make(map[interface{}]interface{})
	unmarshallError := yaml.Unmarshal([]byte(serverconfContent), &serverconfMap)
	if unmarshallError != nil {
		log.Errorf("Error unmarshalling server.conf.yaml: %v", unmarshallError)
		return nil, unmarshallError
	}

	isTrue := true
	keystoreFile := "/home/aceuser/ace-server/https-keystore.p12"
	keystorePassword := "brokerHTTPSKeystore::password"
	keystoreType := "PKCS12"

	ResourceManagersMap := make(map[interface{}]interface{})
	ResourceManagersMap["HTTPSConnector"] = map[string]interface{}{
		"KeystoreFile":     keystoreFile,
		"KeystorePassword": keystorePassword,
		"KeystoreType":     keystoreType,
	}
	// Only update if there is not an existing entry in the override server.conf.yaml
	// so we don't overwrite any customer provided configuration
	if serverconfMap["forceServerHTTPS"] == nil {
		serverconfMap["forceServerHTTPS"] = isTrue
		log.Printf("Force Flows HTTPS Security setting forceServerHTTPS to true")
	}

	if serverconfMap["ResourceManagers"] == nil {
		serverconfMap["ResourceManagers"] = ResourceManagersMap
		log.Printf("Force Flows HTTPS Security creating ResourceManagers->HTTPSConnector")
	} else {
		resourceManagers := serverconfMap["ResourceManagers"].(map[interface{}]interface{})
		if resourceManagers["HTTPSConnector"] == nil {
			resourceManagers["HTTPSConnector"] = ResourceManagersMap["HTTPSConnector"]
			log.Printf("Force Flows HTTPS Security updating ResourceManagers creating HTTPSConnector")
		} else {
			httpsConnector := resourceManagers["HTTPSConnector"].(map[interface{}]interface{})
			log.Printf("Force Flows HTTPS Security merging ResourceManagers->HTTPSConnector")

			if httpsConnector["KeystoreFile"] == nil {
				httpsConnector["KeystoreFile"] = keystoreFile
			} else {
				log.Printf("Force Flows HTTPS Security leaving ResourceManagers->HTTPSConnector->KeystoreFile unchanged")
			}
			if httpsConnector["KeystorePassword"] == nil {
				httpsConnector["KeystorePassword"] = keystorePassword
			} else {
				log.Printf("Force Flows HTTPS Security leaving ResourceManagers->HTTPSConnector->KeystorePassword unchanged")
			}
			if httpsConnector["KeystoreType"] == nil {
				httpsConnector["KeystoreType"] = keystoreType
			} else {
				log.Printf("Force Flows HTTPS Security leaving ResourceManagers->HTTPSConnector->KeystoreType unchanged")
			}
		}
	}

	serverconfYaml, marshallError := yaml.Marshal(&serverconfMap)
	if marshallError != nil {
		log.Errorf("Error marshalling server.conf.yaml: %v", marshallError)
		return nil, marshallError
	}

	return serverconfYaml, nil
}

func generatePassword(length int64) string {
	var i, e = big.NewInt(length), big.NewInt(10)
	bigInt, _ := rand.Int(rand.Reader, i.Exp(e, i, nil))
	return bigInt.String()
}

func watchForceFlowsHTTPSSecret(password string) *fsnotify.Watcher {

	//set up watch on the /home/aceuser/httpsNodeCerts/tls.key file
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("Error creating new watcher for Force Flows to be HTTPS: %v", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// https://github.com/istio/istio/issues/7877 Remove is triggered for the ..data directory when the secret is updated so check for remove too
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("modified file regenerating /home/aceuser/ace-server/https-keystore.p12 and restarting HTTPSConnector:", event.Name)
					time.Sleep(1 * time.Second)

					// 1. generate new p12 /home/aceuser/ace-server/https-keystore.p12 and then
					generateHTTPSKeystore("/home/aceuser/httpsNodeCerts/tls.key", "/home/aceuser/httpsNodeCerts/tls.crt", "/home/aceuser/ace-server/https-keystore.p12", password)

					// 2. patch the HTTPSConnector to pick this up
					patchHTTPSConnector("/home/aceuser/ace-server/config/IntegrationServer.uds")

					// 3. Need to start watching the newly created/ mounted tls.key
					err = watcher.Add("/home/aceuser/httpsNodeCerts/..data/tls.key")
					if err != nil {
						log.Errorf("Error watching /home/aceuser/httpsNodeCerts/tls.key for Force Flows to be HTTPS: %v", err)
					}
				}

			case err, ok := <-watcher.Errors:
				log.Println("error from Force Flows to be HTTPS watcher:", err)
				if !ok {
					return
				}
			}
		}
	}()

	return watcher
}

var generateHTTPSKeystore = localGenerateHTTPSKeystore

func localGenerateHTTPSKeystore(privateKeyLocation string, certificateLocation string, keystoreLocation string, password string) {
	// create /home/aceuser/ace-server/https-keystore.p12 using:
	// single /home/aceuser/httpsNodeCerts/tls.key
	// single /home/aceuser/httpsNodeCerts/tls.crt

	//Script version: openssl pkcs12 -export -in ${certfile} -inkey ${keyfile} -out /home/aceuser/ace-server/https-keystore.p12 -name ${alias} -password pass:${1} 2>&1)

	// Load the private key file into a rsa.PrivateKey
	privateKeyFile, err := ioutil.ReadFile(privateKeyLocation)
	if err != nil {
		log.Error("Error loading "+privateKeyLocation, err)
	}
	privateKeyPem, _ := pem.Decode(privateKeyFile)
	if privateKeyPem.Type != "RSA PRIVATE KEY" {
		log.Error(privateKeyLocation + " is not of type RSA private key")
	}
	privateKeyPemBytes := privateKeyPem.Bytes
	parsedPrivateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPemBytes)
	if err != nil {
		log.Error("Error parsing "+privateKeyLocation+" RSA PRIVATE KEY", err)
	}

	// Load the single cert file into a x509.Certificate
	certificateFile, err := ioutil.ReadFile(certificateLocation)
	if err != nil {
		log.Error("Error loading "+certificateLocation, err)
	}
	certificatePem, _ := pem.Decode(certificateFile)
	if certificatePem.Type != "CERTIFICATE" {
		log.Error(certificateLocation+" is not CERTIFICATE type ", certificatePem.Type)
	}
	certificatePemBytes := certificatePem.Bytes
	parsedCertificate, err := x509.ParseCertificate(certificatePemBytes)
	if err != nil {
		log.Error("Error parsing "+certificateLocation+" CERTIFICATE", err)
	}

	// Create Keystore
	pfxBytes, err := pkcs12.Encode(rand.Reader, parsedPrivateKey, parsedCertificate, []*x509.Certificate{}, password)
	if err != nil {
		log.Error("Error creating the "+keystoreLocation, err)
	}

	// Write out the Keystore 600 (rw- --- ---)
	err = ioutil.WriteFile(keystoreLocation, pfxBytes, 0600)
	if err != nil {
		log.Error(err)
	}
}

var patchHTTPSConnector = localPatchHTTPSConnector

func localPatchHTTPSConnector(uds string) {
	// curl -GET --unix-socket /home/aceuser/ace-server/config/IntegrationServer.uds http://localhost/apiv2/resource-managers/https-connector
	// curl -d "" -POST --unix-socket /home/aceuser/ace-server/config/IntegrationServer.uds http://localhost/apiv2/resource-managers/https-connector/refresh-tls-config -i
	// HTTP/1.1 200 OK
	// Content-Length: 0
	// Content-Type: application/json

	// use unix domain socket
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", uds)
			},
		},
	}

	var err error
	// can be any path root but use localhost to match curl
	_, err = httpc.Post("http://localhost/apiv2/resource-managers/https-connector/refresh-tls-config", "application/octet-stream", strings.NewReader(""))
	if err != nil {
		log.Println("error during call to restart HTTPSConnector for Force Flows HTTPS", err)
	} else {
		log.Println("Call made to restart HTTPSConnector for Force Flows HTTPS")
	}
}
