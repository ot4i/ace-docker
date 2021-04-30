package configuration

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ot4i/ace-docker/common/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
)

const workdirName = "ace-server"
const truststoresName = "truststores"
const keystoresName = "keystores"
const genericName = "generic"
const odbcIniName = "odbc"
const adminsslName = "adminssl"
const aceInstall = "/opt/ibm/ace-11/server/bin"

var (
	configurationClassGVR = schema.GroupVersionResource{
		Group:    "appconnect.ibm.com",
		Version:  "v1beta1",
		Resource: "configurations",
	}

	integrationServerClassGVR = schema.GroupVersionResource{
		Group:    "appconnect.ibm.com",
		Version:  "v1beta1",
		Resource: "integrationservers",
	}
)

/**
* START: FUNCTIONS CREATES EXTERNAL REQUESTS
 */

func getPodNamespace() (string, error) {
	if data, err := ioutilReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
		return "default", err
	}
	return "default", nil
}

func writeConfigurationFile(dir string, fileName string, contents []byte) error {
	makeDirErr := osMkdirAll(dir, 0740)
	if makeDirErr != nil {
		return makeDirErr
	}
	return ioutilWriteFile(dir+string(os.PathSeparator)+fileName, contents, 0740)
}

func unzip(log logger.LoggerInterface, dir string, contents []byte) error {
	var filenames []string
	zipReader, err := zip.NewReader(bytes.NewReader(contents), int64(len(contents)))
	if err != nil {
		log.Printf("%s: %#v", "Failed to read zip contents", err)
		return err
	}

	for _, file := range zipReader.File {

		// Store filename/path for returning and using later on
		filePath := filepath.Join(dir, file.Name)

		// Check for ZipSlip.
		if !strings.HasPrefix(filePath, filepath.Clean(dir)+string(os.PathSeparator)) {
			if err != nil {
				log.Printf("%s: %#v", "Illegal file path:"+filePath, err)
				return err
			}
		}

		filenames = append(filenames, filePath)

		if file.FileInfo().IsDir() {
			// Make Folder
			osMkdirAll(filePath, os.ModePerm)
			continue
		}

		// Make File
		err = osMkdirAll(filepath.Dir(filePath), os.ModePerm)

		if err != nil {
			log.Printf("%s: %#v", "Illegal file path:"+filePath, err)
			return err
		}

		outFile, err := osOpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())

		if err != nil {
			log.Printf("%s: %#v", "Cannot create file writer"+filePath, err)
			return err
		}

		fileReader, err := file.Open()

		if err != nil {
			log.Printf("%s: %#v", "Cannot open file"+filePath, err)
			return err
		}

		_, err = ioCopy(outFile, fileReader)
		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		fileReader.Close()

		if err != nil {
			log.Printf("%s: %#v", "Cannot write file"+filePath, err)
			return err
		}

	}
	return nil
}

/**
* END: FUNCTIONS CREATES EXTERNAL REQUESTS
 */

type configurationObject struct {
	name       string
	configType string
	contents   []byte
}

func getAllConfigurationsImpl(log logger.LoggerInterface, namespace string, configurationsNames []string, dynamicClient dynamic.Interface) ([]*unstructured.Unstructured, error) {

	list := make([]*unstructured.Unstructured, len(configurationsNames))
	for index, configurationName := range configurationsNames {

		res := dynamicClient.Resource(configurationClassGVR).Namespace(namespace)
		configuration, err := res.Get(configurationName, metav1.GetOptions{})
		if err != nil {
			log.Printf("%s: %#v", "Failed to get configuration: "+configurationName, err)
			return nil, err
		}
		list[index] = configuration
	}
	return list, nil
}

var getAllConfigurations = getAllConfigurationsImpl

func getSecretImpl(basedir string, secretName string) ([]byte, error) {
	content, err := ioutil.ReadFile(basedir + string(os.PathSeparator) + "secrets" + string(os.PathSeparator) + secretName + string(os.PathSeparator) + "configuration")
	return content, err
}

var getSecret = getSecretImpl

func parseConfigurationList(log logger.LoggerInterface, basedir string, list []*unstructured.Unstructured) ([]configurationObject, error) {
	output := make([]configurationObject, len(list))
	for index, item := range list {
		name := item.GetName()
		configType, exists, err := unstructured.NestedString(item.Object, "spec", "type")

		if !exists || err != nil {
			log.Printf("%s: %#v", "A configuration must has a type", errors.New("A configuration must has a type"))
			return nil, errors.New("A configuration must has a type")
		}
		switch configType {
		case "policyproject", "odbc", "serverconf":
			fld, exists, err := unstructured.NestedString(item.Object, "spec", "contents")
			if !exists || err != nil {
				log.Printf("%s: %#v", "A configuration with type: "+configType+" must has a contents field", errors.New("A configuration with type: "+configType+" must has a contents field"))
				return nil, errors.New("A configuration with type: " + configType + " must has a contents field")
			}
			contents, err := base64.StdEncoding.DecodeString(fld)
			if err != nil {
				log.Printf("%s: %#v", "Failed to decode contents", err)
				return nil, errors.New("Failed to decode contents")
			}
			output[index] = configurationObject{name: name, configType: configType, contents: contents}
		case "truststorecertificate", "truststore", "keystore", "setdbparms", "generic", "adminssl", "agentx", "agenta", "accounts", "loopbackdatasource":
			if configType == "accounts" {
				designerAuthMode, ok := os.LookupEnv("DEVELOPMENT_MODE")
				if ok && designerAuthMode == "true" {
					log.Println("Ignore accounts.yaml configuration in designer authoring integration server")
					output[index] = configurationObject{name: name, configType: configType, contents: nil}
					break
				}
			}
			secretName, exists, err := unstructured.NestedString(item.Object, "spec", "secretName")
			if !exists || err != nil {
				log.Printf("%s: %#v", "A configuration with type: "+configType+" must have a secretName field", errors.New("A configuration with type: "+configType+" must have a secretName field"))
				return nil, errors.New("A configuration with type: " + configType + " must have a secretName field")
			}
			secretVal, err := getSecret(basedir, secretName)
			if err != nil {
				log.Printf("%s: %#v", "Failed to get secret", err)
				return nil, err
			}
			output[index] = configurationObject{name: name, configType: configType, contents: secretVal}
		}
	}
	return output, nil
}

var dynamicNewForConfig = dynamic.NewForConfig
var kubernetesNewForConfig = kubernetes.NewForConfig

func setupClientsImpl() (dynamic.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamicNewForConfig(config)

	if err != nil {
		return nil, err

	}
	return dynamicClient, nil

}

var setupClients = setupClientsImpl

func SetupConfigurationsFiles(log logger.LoggerInterface, basedir string) error {
	configurationNames, ok := os.LookupEnv("ACE_CONFIGURATIONS")
	if ok && configurationNames != "" {
		log.Printf("Setup configuration files - configuration names: %s", configurationNames)

		return SetupConfigurationsFilesInternal(log, strings.SplitN(configurationNames, ",", -1), basedir)
	} else {
		return nil
	}
}
func SetupConfigurationsFilesInternal(log logger.LoggerInterface, configurationNames []string, basedir string) error {
	// set up k8s client
	dynamicClient, err := setupClients()
	if err != nil {
		return err
	}
	// get pod namespace
	namespace, err := getPodNamespace()
	if err != nil {
		return err
	}
	// get contents for all configurations
	rawConfigurations, err := getAllConfigurations(log, namespace, configurationNames, dynamicClient)

	if err != nil {
		return err
	}
	configurationObjects, err := parseConfigurationList(log, basedir, rawConfigurations)
	if err != nil {
		return err
	}

	for _, configObject := range configurationObjects {
		// create files on the system
		err := constructConfigurationsOnFileSystem(log, basedir, configObject.name, configObject.configType, configObject.contents)
		if err != nil {
			return err
		}
	}
	return nil
}

func constructConfigurationsOnFileSystem(log logger.LoggerInterface, basedir string, configName string, configType string, contents []byte) error {
	log.Printf("Construct a configuration on the filesystem - configuration name: %s type: %s", configName, configType)
	switch configType {
	case "policyproject":
		return constructPolicyProjectOnFileSystem(log, basedir, contents)
	case "truststore":
		return constructTrustStoreOnFileSystem(log, basedir, configName, contents)
	case "keystore":
		return constructKeyStoreOnFileSystem(log, basedir, configName, contents)
	case "odbc":
		return constructOdbcIniOnFileSystem(log, basedir, contents)
	case "serverconf":
		return constructServerConfYamlOnFileSystem(log, basedir, contents)
	case "setdbparms":
		return executeSetDbParms(log, basedir, contents)
	case "generic":
		return constructGenericOnFileSystem(log, basedir, contents)
	case "loopbackdatasource":
		return constructLoopbackDataSourceOnFileSystem(log, basedir, contents)
	case "adminssl":
		return constructAdminSSLOnFileSystem(log, basedir, contents)
	case "accounts":
		designerAuthMode, ok := os.LookupEnv("DEVELOPMENT_MODE")
		if ok && designerAuthMode == "true" {
			// no mount required for this type as already mounted
			return nil
		} else {
			return SetupTechConnectorsConfigurations(log, basedir, contents)
		}
	case "agentx":
		return constructAgentxOnFileSystem(log, basedir, contents)
	case "agenta":
		return constructAgentaOnFileSystem(log, basedir, contents)
	case "truststorecertificate":
		return addTrustCertificateToCAcerts(log, basedir, configName, contents)
	default:
		return errors.New("Unknown configuration type")
	}
}

func constructPolicyProjectOnFileSystem(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Construct policy project on the filesystem")
	return unzip(log, basedir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"overrides", contents)
}

func constructTrustStoreOnFileSystem(log logger.LoggerInterface, basedir string, name string, contents []byte) error {
	log.Printf("Construct truststore on the filesystem - Truststore name: %s", name)
	return writeConfigurationFile(basedir+string(os.PathSeparator)+truststoresName, name, contents)
}

func constructKeyStoreOnFileSystem(log logger.LoggerInterface, basedir string, name string, contents []byte) error {
	log.Printf("Construct keystore on the filesystem - Keystore name: %s", name)
	return writeConfigurationFile(basedir+string(os.PathSeparator)+keystoresName, name, contents)
}

func constructOdbcIniOnFileSystem(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Construct odbc.Ini on the filesystem")
	return writeConfigurationFile(basedir+string(os.PathSeparator)+workdirName, "odbc.ini", contents)
}

func constructGenericOnFileSystem(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Construct generic files on the filesystem")
	return unzip(log, basedir+string(os.PathSeparator)+genericName, contents)
}

func constructLoopbackDataSourceOnFileSystem(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Construct loopback connector files on the filesystem")
	return unzip(log, basedir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config"+string(os.PathSeparator)+"connectors"+string(os.PathSeparator)+"loopback", contents)
}

func constructAdminSSLOnFileSystem(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Construct adminssl on the filesystem")
	return unzip(log, basedir+string(os.PathSeparator)+adminsslName, contents)
}

func constructServerConfYamlOnFileSystem(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Construct serverconfyaml on the filesystem")
	return writeConfigurationFile(basedir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"overrides", "server.conf.yaml", contents)
}

func constructAgentxOnFileSystem(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Construct agentx on the filesystem")
	return writeConfigurationFile(basedir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config/iibswitch/agentx", "agentx.json", contents)
}

func constructAgentaOnFileSystem(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Construct agenta on the filesystem")
	return writeConfigurationFile(basedir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config/iibswitch/agenta", "agenta.json", contents)
}

func addTrustCertificateToCAcerts(log logger.LoggerInterface, basedir string, name string, contents []byte) error {
	log.Println("Adding trust certificate to CAcerts")
	// creating temporary file based on the content
	tmpFile := creatingTempFile(log, contents, name)
	// cleans up the file afterwards
	defer os.Remove(tmpFile.Name())
	// adding this file to CAcerts
	commandCreateArgsJKS := []string{"-import", "-file", tmpFile.Name(), "-alias", name, "-keystore", "$MQSI_JREPATH/lib/security/cacerts", "-storepass", "changeit", "-noprompt", "-storetype", "JKS"}
	return internalRunKeytoolCommand(log, commandCreateArgsJKS)
}

func creatingTempFile(log logger.LoggerInterface, contents []byte, name string) *os.File {
	tmpFile, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		log.Println("Cannot create temporary file", err)
	}

	// writing content to the file
	if _, err = tmpFile.Write(contents); err != nil {
		log.Println("Failed to write to temporary file", err)
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		log.Println("Failed to close the file", err)
	}
	return tmpFile
}

func executeSetDbParms(log logger.LoggerInterface, basedir string, contents []byte) error {
	log.Println("Execute mqsisetdbparms command")
	for index, m := range strings.Split(string(contents), "\n") {
		// ignore empty lines
		if len(strings.TrimSpace(m)) > 0 {
			contentsArray := strings.Fields(strings.TrimSpace(m))
			log.Printf("Execute line %d with number of args: %d", index, len(contentsArray))
			var trimmedArray []string
			for _, m := range contentsArray {
				escapedQuote := strings.Replace(m, "'", "'\\''", -1)
				trimmedArray = append(trimmedArray, "'"+strings.TrimSpace(escapedQuote)+"'")
			}
			if len(trimmedArray) > 2 {
				if trimmedArray[0] == "'mqsisetdbparms'" {
					if !Contains(trimmedArray, "'-w'") {
						trimmedArray = append(trimmedArray, "'-w'")
						trimmedArray = append(trimmedArray, "'"+basedir+string(os.PathSeparator)+workdirName+"'")
					}
					err := internalRunSetdbparmsCommand(log, "mqsisetdbparms", trimmedArray[1:])
					if err != nil {
						return err
					}
				} else if len(trimmedArray) == 3 {
					args := []string{"'-n'", trimmedArray[0], "'-u'", trimmedArray[1], "'-p'", trimmedArray[2], "'-w'", "'" + basedir + string(os.PathSeparator) + workdirName + "'"}
					err := internalRunSetdbparmsCommand(log, "mqsisetdbparms", args)
					if err != nil {
						return err
					}
				} else {
					return errors.New("Invalid mqsisetdbparms entry - too many parameters")
				}
			} else {
				return errors.New("Invalid mqsisetdbparms entry - too few parameters")
			}
		}
	}
	return nil

}
func runSetdbparmsCommand(log logger.LoggerInterface, command string, params []string) error {
	realCommand := command
	return runCommand(log, realCommand, params)
}

func runCommand(log logger.LoggerInterface, command string, params []string) error {
	realCommand := "source " + aceInstall + "/mqsiprofile && " + command + " " + strings.Join(params[:], " ")
	cmd := exec.Command("/bin/sh", "-c", realCommand)
	cmd.Stdin = strings.NewReader("some input")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		log.Printf("Error executing command: %s %s", stdout.String(), stderr.String())
	} else {
		log.Printf("Successfully executed command.")
	}
	return err

}

func runKeytoolCommand(log logger.LoggerInterface, params []string) error {
	return runCommand(log, "keytool", params)

}

var internalRunSetdbparmsCommand = runSetdbparmsCommand
var internalRunKeytoolCommand = runKeytoolCommand

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
func main() {
}
