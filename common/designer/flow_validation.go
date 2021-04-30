/*
© Copyright IBM Corporation 2020

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

// Package designer contains code for the designer specific logic
package designer

import (
	"os"
	"io"
	"io/ioutil"
	"strings"
	"gopkg.in/yaml.v2"

	"github.com/ot4i/ace-docker/internal/command"
	"github.com/ot4i/ace-docker/common/logger"
)

var runAsUser = command.RunAsUser
var osOpen = os.Open
var osCreate = os.Create
var ioCopy = io.Copy
var rename = os.Rename
var removeAll = os.RemoveAll
var remove = os.Remove
var mkdir = os.Mkdir
var readDir = ioutil.ReadDir

type flowDocument struct {
	Integration integration `yaml:"integration"`
}

type integration struct {
	TriggerInterfaces map[string]flowInterface `yaml:"trigger-interfaces"`
	ActionInterfaces map[string]flowInterface `yaml:"action-interfaces"`
}

type flowInterface struct {
	ConnectorType string `yaml:"connector-type"`
}

// replaceFlow replaces the resources associated with a deployed designer flow
// with the unpacked generic invalid BAR file
var replaceFlow = func (flow string, log logger.LoggerInterface, basedir string) error {
	runFolder := basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" 

	// Delete the /run/<flow>PolicyProject folder
	err := removeAll(runFolder + string(os.PathSeparator) + flow + "PolicyProject")
	if err != nil {
		log.Errorf("Error checking for the flow's policy project folder: %v", err)
		return err
	}	

	// Rename the flow folder <flow>_invalid_license
	newFlowName := flow + "_invalid_license"
	err = rename(runFolder + string(os.PathSeparator) + flow, runFolder + string(os.PathSeparator) + newFlowName)
	if err != nil {
		log.Errorf("Error renaming the flow's folder: %v", err)
		return err
	}
	flow = newFlowName

	// Delete the /run/<flow>/gen folder
	err = removeAll(runFolder + string(os.PathSeparator) + flow + string(os.PathSeparator) + "gen")
	if err != nil {
		log.Errorf("Error deleting the flow's /gen folder: %v", err)
		return err
	}

	// Create a new /run/<flow>/gen folder
	err = mkdir(runFolder + string(os.PathSeparator) + flow + string(os.PathSeparator) + "gen", 0777)
	if err != nil {
		log.Errorf("Error creating the flow's new /gen folder: %v", err)
		return err
	}

	tempFolder := basedir + string(os.PathSeparator) + "temp"
	// Copy the contents of /temp/gen into /run/<flow>/gen
	invalidFlowResourcesList, err := readDir(tempFolder + string(os.PathSeparator) + "gen")
	if err != nil {
		log.Errorf("Error checking for the invalid flow's /gen folder: %v", err)
		return err
	}
	for _, invalidFlowResource := range invalidFlowResourcesList {
		err = copy(tempFolder + string(os.PathSeparator) + "gen" + string(os.PathSeparator) + invalidFlowResource.Name(), runFolder + string(os.PathSeparator) + flow + string(os.PathSeparator) + "gen" + string(os.PathSeparator) + invalidFlowResource.Name(), log)
		if err != nil {
			log.Errorf("Error copying resource %v from the flow's /gen folder: %v", invalidFlowResource.Name(), err)
			return err
		}
	}

	// Remove the .msgflow, .subflow and, restapi.descriptor from /run/<flow>
	flowResourcesList, err := readDir(runFolder + string(os.PathSeparator) + flow)
	if err != nil {
		log.Errorf("Error checking for the flow's folder: %v", err)
		return err
	}
	for _, flowResource := range flowResourcesList {
		if strings.Contains(flowResource.Name(), ".msgflow") || strings.Contains(flowResource.Name(), ".subflow") || flowResource.Name() == "restapi.descriptor" {
			err = remove(runFolder+ string(os.PathSeparator) + flow + string(os.PathSeparator) + flowResource.Name())
			if err != nil {
				log.Errorf("Error deleting resource %v from the flow's folder: %v", flowResource.Name(), err)
				return err
			}
		}
	}

	// Replace restapi.descriptor with application.descriptor
	err = copy(tempFolder + string(os.PathSeparator) + "application.descriptor", runFolder + string(os.PathSeparator) + flow + string(os.PathSeparator) + "application.descriptor", log)
	if err != nil {
		log.Errorf("Error copying resource restapi.descriptor to application.descriptor: %v", err)
		return err
	}
	return nil
}

// cleanupInvalidBarResources deletes the unpacked generic invalid BAR file
// to ensure there's no unknown flows deployed to a user's instnace
func cleanupInvalidBarResources(basedir string) error {
	return os.RemoveAll(basedir + string(os.PathSeparator) + "temp")
}

// getConnectorLicenseToggle computes the license toggle name from the connector
func getConnectorLicenseToggleName(name string) string {
	return "connector-" + name
}

// findDisabledConnectorInFlow returns the first disabled connector it finds
// if it doesn't find a disabled connector, it returns an empty string
var findDisabledConnectorInFlow = func (flowDocument flowDocument, log logger.LoggerInterface) string {
	disabledConnectors := make([]string, 0)

	// read the connector-type field under each interface
	// and check if the license toggle for that connector is enabled
	findDisabledConnector := func(interfaces map[string]flowInterface) {
		for _, i := range interfaces {
			connector := i.ConnectorType
			if connector != "" {
				log.Printf("Checking if connector %v is supported under the current license.", connector)
				if !isLicenseToggleEnabled(getConnectorLicenseToggleName(connector)) {
					disabledConnectors = append(disabledConnectors, connector)
				}
			}
		}
	}

	findDisabledConnector(flowDocument.Integration.TriggerInterfaces)
	findDisabledConnector(flowDocument.Integration.ActionInterfaces)

	return strings.Join(disabledConnectors[:], ", ")
}

// IsFlowValid checks if a single flow is valid
var IsFlowValid = func(log logger.LoggerInterface, flow string, flowFile []byte) (bool, error) {
	var flowDocument flowDocument
	err := yaml.Unmarshal(flowFile, &flowDocument)
	if err != nil {
		log.Errorf("Error processing running flow in folder %v: %v", flow, err)
		return false, err
	}

	disabledConnectors := findDisabledConnectorInFlow(flowDocument, log)
	if disabledConnectors != "" {
		log.Errorf("Flow %v contains one or more connectors that aren't supported under the current license. Please update your license to enable this flow to run. The unsupported connectors are: %v.", flow, disabledConnectors)
	}
	return disabledConnectors == "", nil
}

// ValidateFlows checks if the flows in the /run directory are valid
// by making sure all the connectors used by a flow are supported under the license
// if invalid, then it replaces a flow with a generic invalid one, which fails at startup
// if valid, it doesn't do anything
func ValidateFlows(log logger.LoggerInterface, basedir string) error {
	// at this point the /run directory should have been created
	log.Println("Processing running flows in folder /home/aceuser/ace-server/run")
	runFileList, err := ioutil.ReadDir(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run")
	if err != nil {
		log.Errorf("Error checking for the /run folder: %v", err)
		return err
	}

	for _, file := range runFileList {
		// inside the /run directory there are folders corresponding to running toolkit and designer flows
		// designer flows will have two folders: /run/<flow name> and /run/<flow name>PolicyProject
		// the yaml of the designer flow is the /run/<flow name> folder
		flow := file.Name()
		if file.IsDir() && !strings.Contains(flow, "PolicyProject") && dirExists(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + "PolicyProject") {
			log.Printf("Processing running flow with name %v", flow)
			flowFile, err := ioutil.ReadFile(basedir + string(os.PathSeparator) + "ace-server" + string(os.PathSeparator) + "run" + string(os.PathSeparator) + flow + string(os.PathSeparator) + flow + ".yaml")
			if err != nil {
				log.Errorf("Error processing running flow in folder %v: %v", flow, err)
				return err
			}

			valid, err := IsFlowValid(log, flow, flowFile)
			if err != nil {
				return err
			}
			if !valid {
				err = replaceFlow(flow, log, basedir)
				if err != nil {
					log.Errorf("Error replacing running flow in folder %v: %v", flow, err)
					return err
				}
			}
		}
	}

	return cleanupInvalidBarResources(basedir)
}

// copy copies from src to dest
var copy = func (src string, dest string, log logger.LoggerInterface) error {
	source, err := osOpen(src)
	if err != nil {
		log.Errorf("Error opening the source file %v: %v", src, err)
		return err
	}
	defer source.Close()

	destination, err := osCreate(dest)
	if err != nil {
		log.Errorf("Error creating the destination file %v: %v", dest, err)
		return err
	}
	defer destination.Close()
	_, err = ioCopy(destination, source)
	return err
}

// dirExist cheks if a directory exists at the given path
var dirExists = func (path string) bool {
	_, err := os.Stat(path)
	if err == nil { 
		return true 
	}
	return !os.IsNotExist(err)
}
