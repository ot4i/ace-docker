package iscommandsapi

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ot4i/ace-docker/common/logger"
	"github.com/stretchr/testify/assert"
)

var startHTTPServerRestore = startHTTPServer
var restartIntegrationServerFuncRestore = restartIntegrationServerFunc
var runCommandRestore func(name string, args ...string) (string, int, error) = runCommand

type testCommandHandler struct {
	executeHandler func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError)
}

func (me *testCommandHandler) execute(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {
	return me.executeHandler(log, body)
}

func reset() {
	startHTTPServer = func(log logger.LoggerInterface, portNo int) *http.Server {
		panic("Start http server should be mocked")
	}

	restartIntegrationServerFunc = func() error {
		panic("Restart integration server should be mocked")
	}

	runCommand = func(name string, args ...string) (string, int, error) {
		panic("Run command should be implemented")
	}
}

func restore() {
	runCommand = runCommandRestore
	startHTTPServer = startHTTPServerRestore
	restartIntegrationServerFunc = restartIntegrationServerFuncRestore
}

func TestStartCommandsAPIServer(t *testing.T) {

	var testLogger, _ = logger.NewLogger(os.Stdout, true, true, "test")
	restartIsFunc := func() error { return nil }

	t.Run("When restart func is nil, returns error", func(t *testing.T) {
		reset()
		defer restore()

		err := StartCommandsAPIServer(testLogger, 123, nil)

		assert.NotNil(t, err)
		assert.Equal(t, "Restart handler should not be nil", err.Error())
	})

	t.Run("registers setdbparms command handler and invokes startHTTPServer with specified portNo", func(t *testing.T) {
		reset()
		defer restore()

		var logP logger.LoggerInterface
		var httpServerInstance = &http.Server{}
		var portNoP int
		startHTTPServer = func(log logger.LoggerInterface, portNo int) *http.Server {
			logP = log
			portNoP = portNo

			return httpServerInstance
		}

		err := StartCommandsAPIServer(testLogger, 123, restartIsFunc)

		assert.NotNil(t, commandsHandler["setdbparms"])
		assert.Nil(t, err)
		assert.Equal(t, logP, testLogger)
		assert.Equal(t, portNoP, 123)
		assert.Equal(t, httpServerInstance, httpServer)

	})
}

func TestCommandRequestHttpHandler(t *testing.T) {

	var request *http.Request
	var response *httptest.ResponseRecorder
	var handler http.HandlerFunc
	var testCommandURL string = "/commands/test"

	setupTestCommand := func(requestURL string, executeCommandFunc func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError)) {
		restartIntegrationServerFunc = func() error {
			return nil
		}

		handler = http.HandlerFunc(commandRequestHandler)
		request, _ = http.NewRequest("POST", requestURL, nil)
		response = httptest.NewRecorder()

		testCommandHadler := testCommandHandler{executeHandler: executeCommandFunc}
		handleCRUDCommand("test", &testCommandHadler)
	}

	t.Run("Returns not found when url doesn't have command ", func(t *testing.T) {
		reset()
		setupTestCommand("/commands", nil)
		defer restore()

		handler.ServeHTTP(response, request)
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, `{"success":false,"message":"Not found"}`+"\n", response.Body.String())
	})

	t.Run("when no handler registered for the command, returns not found", func(t *testing.T) {
		reset()
		defer restore()

		setupTestCommand("/commands/test2", nil)

		handler.ServeHTTP(response, request)
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, `{"success":false,"message":"Not found"}`+"\n", response.Body.String())
	})

	t.Run("Invokes registered command handlers with request body", func(t *testing.T) {
		reset()
		defer restore()

		called := false
		executeCommandFunc := func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {
			called = true
			return &commandResponse{message: "ok"}, nil
		}

		setupTestCommand(testCommandURL, executeCommandFunc)

		handler.ServeHTTP(response, request)

		assert.Equal(t, true, called)
	})

	t.Run("when command handler returns message, responds with status ok and message", func(t *testing.T) {
		reset()
		defer restore()

		executeCommandFunc := func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {
			return &commandResponse{message: "command executed", restartIs: true}, nil
		}

		setupTestCommand(testCommandURL, executeCommandFunc)

		handler.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, `{"success":true,"message":"command executed"}`+"\n", response.Body.String())
	})

	t.Run("when command handler returns invalid input, responds with bad request", func(t *testing.T) {
		reset()
		defer restore()

		executeCommandFunc := func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {
			return nil, &commandError{error: "Invalid input", errorCode: commandErrorInvalidInput}
		}

		setupTestCommand(testCommandURL, executeCommandFunc)

		handler.ServeHTTP(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, `{"success":false,"message":"Invalid input"}`+"\n", response.Body.String())
	})

	t.Run("when command handler returns internal error, responds with internal server error", func(t *testing.T) {
		reset()
		defer restore()

		executeCommandFunc := func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {
			return nil, &commandError{error: "Internal error while invkoing command", errorCode: commandErrorInternal}
		}

		setupTestCommand(testCommandURL, executeCommandFunc)

		handler.ServeHTTP(response, request)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, `{"success":false,"message":"Internal error while invkoing command"}`+"\n", response.Body.String())
	})

	t.Run("when command handler returns restartIs with true in response, invokes restart integration server func", func(t *testing.T) {
		reset()
		defer restore()

		executeCommandFunc := func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {
			return &commandResponse{message: "command executed", restartIs: true}, nil
		}

		setupTestCommand(testCommandURL, executeCommandFunc)
		restartIsCalled := false

		restartIntegrationServerFunc = func() error {
			restartIsCalled = true
			return nil
		}

		handler.ServeHTTP(response, request)

		assert.Equal(t, true, restartIsCalled)
	})

	t.Run("when restart inegration server failed, returns internal server error", func(t *testing.T) {
		reset()
		defer restore()

		executeCommandFunc := func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {
			return &commandResponse{message: "command executed", restartIs: true}, nil
		}
		setupTestCommand(testCommandURL, executeCommandFunc)

		restartIsCalled := false
		restartIntegrationServerFunc = func() error {
			restartIsCalled = true
			return errors.New("restart failed")
		}

		handler.ServeHTTP(response, request)

		assert.Equal(t, true, restartIsCalled)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, `{"success":false,"message":"Integration server restart failed"}`+"\n", response.Body.String())
	})

	t.Run("when restart inegration server is nill, returns integration server has not restarted message", func(t *testing.T) {
		reset()
		defer restore()
		executeCommandFunc := func(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError) {
			return &commandResponse{message: "command executed", restartIs: true}, nil
		}

		setupTestCommand(testCommandURL, executeCommandFunc)
		restartIntegrationServerFunc = nil
		handler.ServeHTTP(response, request)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, `{"success":false,"message":"Integration server has not restarted"}`+"\n", response.Body.String())
	})
}
