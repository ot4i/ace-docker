package iscommandsapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/ot4i/ace-docker/common/logger"
	"github.com/ot4i/ace-docker/internal/command"
)

// DbParamsCommand for mqsisetdbparams command
type DbParamsCommand struct {
	ResourceType string `json:"resourceType"`
	ResourceName string `json:"resourceName"`
	UserName     string `json:"userName"`
	Password     string `json:"password"`
}

// DbParamsHandler handler
type DbParamsHandler struct {
}

var runCommand func(name string, args ...string) (string, int, error) = command.Run

func (handler DbParamsHandler) execute(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {

	dbParamsCommand := DbParamsCommand{}
	err := json.NewDecoder(body).Decode(&dbParamsCommand)

	if err != nil {
		log.Println("#setdbparamsHandler Error in decoding json")
		return nil, &commandError{errorCode: commandErrorInvalidInput, error: "Invalid request"}
	}

	if len(dbParamsCommand.UserName) == 0 || len(dbParamsCommand.ResourceType) == 0 || len(dbParamsCommand.ResourceName) == 0 {
		log.Println("#setdbparamsHandler one of required parameters not found")
		return nil, &commandError{errorCode: commandErrorInvalidInput, error: "Invalid request"}
	}

	resource := fmt.Sprintf("%s::%s", dbParamsCommand.ResourceType, dbParamsCommand.ResourceName)

	isCredentialsExists, err := isResourceCredentialsExists(resource, dbParamsCommand.UserName, dbParamsCommand.Password)

	if err != nil {
		log.Printf("#setdbparamsHandler, an error occurred while checking the resource already exists, error %s", err.Error())
		return nil, &commandError{errorCode: commandErrorInternal, error: "Internal error"}
	}

	if isCredentialsExists {
		log.Printf("#setdbparamsHandler credentials with the same user name and password already exists for the resource %s", resource)
		return &commandResponse{message: "success", restartIs: false}, nil
	}

	log.Printf("#setdbparamsHandler adding %s of type %s", dbParamsCommand.ResourceName, dbParamsCommand.ResourceType)

	runCommand("ace_mqsicommand.sh", "setdbparms", "-w", "/home/aceuser/ace-server/", "-n", resource,
		"-u", dbParamsCommand.UserName, "-p", dbParamsCommand.Password)

	return &commandResponse{message: "success", restartIs: true}, nil
}

func isResourceCredentialsExists(resourceName string, username string, password string) (bool, error) {

	cmdOutput, exitCode, err := runCommand("ace_mqsicommand.sh", "reportdbparms", "-w", "/home/aceuser/ace-server/", "-n", resourceName,
		"-u", username, "-p", password)

	if err != nil {
		return false, err
	}

	if exitCode != 0 {
		errorMsg := fmt.Sprintf("mqsireportdbparams command exited with non zero, exit code %v", exitCode)
		return false, errors.New(errorMsg)
	}

	credentialsMatchRegx := regexp.MustCompile("(?m)^(\\s)*\\b(BIP8201I).*\\b(correct).*$")

	return credentialsMatchRegx.MatchString(cmdOutput), nil

}
