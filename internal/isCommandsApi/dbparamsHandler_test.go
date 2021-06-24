package iscommandsapi

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ot4i/ace-docker/common/logger"
)

func TestDbParamsHandlerExecute(t *testing.T) {

	testLogger, _ := logger.NewLogger(os.Stdout, true, true, "test")
	t.Run("When invalid json input is given, returns invalid input", func(t *testing.T) {

		reset()
		defer restore()

		invalidJSON := `{"ResourceName":"abc}`

		cmdResponse, cmdError := DbParamsHandler{}.execute(testLogger, strings.NewReader(invalidJSON))

		assert.NotNil(t, cmdError)
		assert.Equal(t, commandErrorInvalidInput, cmdError.errorCode)
		assert.Equal(t, "Invalid request", cmdError.error)
		assert.Nil(t, cmdResponse)
	})

	t.Run("When resource name is missing, returns invalid input", func(t *testing.T) {

		reset()
		defer restore()

		invalidJSON := `{"ResourceType":"mq","UserName":"abc","Password":"xyz"}`

		cmdResponse, cmdError := DbParamsHandler{}.execute(testLogger, strings.NewReader(invalidJSON))

		assert.NotNil(t, cmdError)
		assert.Equal(t, commandErrorInvalidInput, cmdError.errorCode)
		assert.Equal(t, "Invalid request", cmdError.error)
		assert.Nil(t, cmdResponse)
	})

	t.Run("When resource type is empty, returns invalid input", func(t *testing.T) {

		reset()
		defer restore()

		invalidJSON := `{"resourceType":"","resourceName":"123","UserName":"abc","Password":"xyz"}`

		cmdResponse, cmdError := DbParamsHandler{}.execute(testLogger, strings.NewReader(invalidJSON))

		assert.NotNil(t, cmdError)
		assert.Equal(t, commandErrorInvalidInput, cmdError.errorCode)
		assert.Equal(t, "Invalid request", cmdError.error)
		assert.Nil(t, cmdResponse)
	})

	t.Run("When username  is empty, returns invalid input", func(t *testing.T) {

		reset()
		defer restore()

		invalidJSON := `{"resourceType":"mq","resourceName":"123","UserName":"","Password":"xyz"}`

		cmdResponse, cmdError := DbParamsHandler{}.execute(testLogger, strings.NewReader(invalidJSON))

		assert.NotNil(t, cmdError)
		assert.Equal(t, commandErrorInvalidInput, cmdError.errorCode)
		assert.Equal(t, "Invalid request", cmdError.error)
		assert.Nil(t, cmdResponse)
	})

	t.Run("when request is valid invokes reportdbparms command to check supplied resource credentials", func(t *testing.T) {

		reset()
		defer restore()

		var invokedCommandName string = ""
		var invokedCommandArgs []string = nil

		runCommand = func(cmdName string, cmdArguments ...string) (string, int, error) {

			invokedCommandName = cmdName
			invokedCommandArgs = cmdArguments
			return "", 1, nil
		}

		validJSON := `{"resourceType":"mq","resourceName":"123","UserName":"abc","Password":"xyz"}`
		expectedCommandArgs := []string{"reportdbparms", "-w", "/home/aceuser/ace-server/", "-n", "mq::123", "-u", "abc", "-p", "xyz"}

		_, err := DbParamsHandler{}.execute(testLogger, strings.NewReader(validJSON))

		assert.NotNil(t, err)
		assert.Equal(t, "ace_mqsicommand.sh", invokedCommandName)
		assert.Equal(t, expectedCommandArgs, invokedCommandArgs)
	})

	t.Run("when reportdbparms cmd output contains credentials correct line, returns success with restart integration server flag false", func(t *testing.T) {

		reset()
		defer restore()

		var credentialsExistsOutput = `BIP8180I: The resource name 'mq::123' has userID 'abc'.
		BIP8201I: The password you entered, 'xyz' for resource 'mq::123' and userId 'abc' is correct.`

		runCommand = func(cmdName string, cmdArguments ...string) (string, int, error) {
			return credentialsExistsOutput, 0, nil
		}

		validJSON := `{"resourceType":"mq","resourceName":"123","UserName":"abc","Password":"xyz"}`

		cmdResponse, _ := DbParamsHandler{}.execute(testLogger, strings.NewReader(validJSON))

		assert.NotNil(t, cmdResponse)
		assert.Equal(t, "success", cmdResponse.message)
		assert.Equal(t, false, cmdResponse.restartIs)
	})

	t.Run("When reportdbparams command output doesn't contain credentials correct line, invokes setdbparams and returns success", func(t *testing.T) {

		reset()
		defer restore()

		var invokedCommandName string = ""
		var invokedCommandArgs []string = nil

		reportDbPramsOutput := `BIP8180I: The resource name 'mq::123' has userID 'abc'.
		BIP8204W: The password you entered, 'xyz' for resource 'mq::123' and userId 'test1' is incorrect`

		runCommand = func(cmdName string, cmdArguments ...string) (string, int, error) {

			if cmdArguments[0] == "reportdbparms" {
				return reportDbPramsOutput, 0, nil
			}

			invokedCommandName = cmdName
			invokedCommandArgs = cmdArguments
			return "", 0, nil
		}

		validJSON := `{"resourceType":"mq","resourceName":"123","UserName":"abc","Password":"xyz"}`

		expectedCommandArgs := []string{"setdbparms", "-w", "/home/aceuser/ace-server/", "-n", "mq::123", "-u", "abc", "-p", "xyz"}

		cmdResponse, _ := DbParamsHandler{}.execute(testLogger, strings.NewReader(validJSON))

		assert.NotNil(t, cmdResponse)
		assert.Equal(t, "success", cmdResponse.message)
		assert.Equal(t, "ace_mqsicommand.sh", invokedCommandName)
		assert.Equal(t, expectedCommandArgs, invokedCommandArgs)
	})

	t.Run("When reportdbparams command returns error, returns internal error", func(t *testing.T) {
		reset()
		defer restore()

		runCommand = func(cmdName string, cmdArguments ...string) (string, int, error) {
			if cmdArguments[0] == "reportdbparms" {
				return "", 0, errors.New("some error")
			}

			panic("should not come here")
		}

		validJSON := `{"resourceType":"mq","resourceName":"123","UserName":"abc","Password":"xyz"}`

		cmdResponse, cmdError := DbParamsHandler{}.execute(testLogger, strings.NewReader(validJSON))

		assert.Nil(t, cmdResponse)
		assert.NotNil(t, commandErrorInternal, cmdError.errorCode)
		assert.Equal(t, "Internal error", cmdError.error)
	})

	t.Run("When reportdbparams command exits with non zero, returns internal error", func(t *testing.T) {
		reset()
		defer restore()

		runCommand = func(cmdName string, cmdArguments ...string) (string, int, error) {
			if cmdArguments[0] == "reportdbparms" {
				return "", 1, nil
			}

			return "", 0, nil
		}

		validJSON := `{"resourceType":"mq","resourceName":"123","UserName":"abc","Password":"xyz"}`

		cmdResponse, cmdError := DbParamsHandler{}.execute(testLogger, strings.NewReader(validJSON))

		assert.Nil(t, cmdResponse)
		assert.Equal(t, commandErrorInternal, cmdError.errorCode)
		assert.Equal(t, "Internal error", cmdError.error)
	})
}
