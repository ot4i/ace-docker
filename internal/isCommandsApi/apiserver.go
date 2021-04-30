package iscommandsapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/ot4i/ace-docker/common/logger"
)

var log logger.LoggerInterface

type commandResponse struct {
	message   string
	restartIs bool
}

type commandError struct {
	error     string
	errorCode int
}

const commandErrorInternal = 1
const commandErrorInvalidInput = 2

type commandHandlerInterface interface {
	execute(log logger.LoggerInterface, body io.Reader) (*commandResponse, *commandError)
}

type handlerInvocationDetails struct {
	handler commandHandlerInterface
}

type apiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var commandsHandler map[string]handlerInvocationDetails = make(map[string]handlerInvocationDetails)
var restartIntegrationServerFunc func() error
var httpServer *http.Server = nil

func writeRequestesponse(writer http.ResponseWriter, statusCode int, message string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	apiReturn := apiResponse{}
	apiReturn.Success = statusCode == http.StatusOK
	apiReturn.Message = message

	json.NewEncoder(writer).Encode(&apiReturn)
}

func commandRequestHandler(writer http.ResponseWriter, request *http.Request) {
	url := request.URL.Path

	log.Printf("#commandRequestHandler serving %f, method %s", url, request.Method)
	apiRegEx := regexp.MustCompile("^/commands/(\\w*)/?(.*)?$")

	matches := apiRegEx.FindStringSubmatch(url)

	var err error = nil
	if len(matches) != 3 {
		log.Printf("#commandRequestHandler url doesn't have expected 3 tokens", url)
		writeRequestesponse(writer, http.StatusNotFound, "Not found")
		return
	}

	command := matches[1]
	commandHandler, found := commandsHandler[command]
	if !found {
		log.Printf("#commandRequestHandler No hanlders found for %s", command)
		writeRequestesponse(writer, http.StatusNotFound, "Not found")
		return
	}

	restartIs := false

	apiResponse := ""

	switch request.Method {
	case http.MethodPost:
		commandResponse, commandError := commandHandler.handler.execute(log, request.Body)

		if commandError != nil {
			log.Printf("#commandRequestHandler  an error occurred while processing command %s, error %s", command, commandError.error)

			if commandError.errorCode == commandErrorInvalidInput {
				writeRequestesponse(writer, http.StatusBadRequest, commandError.error)
			} else {
				writeRequestesponse(writer, http.StatusInternalServerError, commandError.error)
			}

			return
		}

		apiResponse = commandResponse.message
		restartIs = commandResponse.restartIs
		break

	default:
		writeRequestesponse(writer, http.StatusNotImplemented, "Not implemented")
		return
	}

	if restartIs {
		log.Printf("#commandRequestHandler command %s requested integration server restart, invoking restart callback", command)
		if restartIntegrationServerFunc != nil {
			err = restartIntegrationServerFunc()

			if err != nil {
				log.Errorf("#commandRequestHandler an error occurred while restarting integration server %v", err)
				apiResponse = "Integration server restart failed"
			}
		} else {
			log.Error("Intergration server function is nil")
			err = errors.New("Integration server has not restarted")
			apiResponse = "Integration server has not restarted"
		}
	}

	if err == nil {
		writeRequestesponse(writer, http.StatusOK, apiResponse)
	} else {
		writeRequestesponse(writer, http.StatusInternalServerError, apiResponse)
	}
}

func handleCRUDCommand(command string, crudCommandHandler commandHandlerInterface) {

	crudHandler := handlerInvocationDetails{
		handler: crudCommandHandler}

	commandsHandler[command] = crudHandler
}

var startHTTPServer = func(logger logger.LoggerInterface, portNo int) *http.Server {
	log = logger

	address := fmt.Sprintf(":%v", portNo)
	server := &http.Server{Addr: address}

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("Error in serving " + err.Error())
		}
		log.Println("Commands API server stopped ")
	}()

	return server
}

// StartCommandsAPIServer Starts the api server
func StartCommandsAPIServer(logger logger.LoggerInterface, portNumber int, restartIsFunc func() error) error {

	if restartIsFunc == nil {
		return errors.New("Restart handler should not be nil")
	}

	log = logger
	restartIntegrationServerFunc = restartIsFunc

	dbParamsHandler := DbParamsHandler{}

	handleCRUDCommand("setdbparms", dbParamsHandler)
	http.HandleFunc("/commands/", commandRequestHandler)
	httpServer = startHTTPServer(log, portNumber)

	return nil
}

// StopCommandsAPIServer Stops commands api server
func StopCommandsAPIServer() {
	if httpServer != nil {
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
		}()

		if err := httpServer.Shutdown(ctxShutDown); err != nil {
			log.Fatalf("Server Shutdown Failed:%+s", err)
		} else {
			log.Println("Integration commands API stopped")
		}
	}

}
