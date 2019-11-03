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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ot4i/ace-docker/internal/logger"
)

var log *logger.Logger

func logTerminationf(format string, args ...interface{}) {
	logTermination(fmt.Sprintf(format, args...))
}

func logTermination(args ...interface{}) {
	msg := fmt.Sprint(args...)
	// Write the message to the termination log.  This is the default place
	// that Kubernetes will look for termination information.
	log.Debugf("Writing termination message: %v", msg)
	err := ioutil.WriteFile("/dev/termination-log", []byte(msg), 0660)
	if err != nil {
		log.Debug(err)
	}
	log.Error(msg)
}

func getLogFormat() string {
	return os.Getenv("LOG_FORMAT")
}

func getLogOutputFormat() string {
	switch getLogFormat() {
	case "json":
		return "ibmjson"
	case "basic":
		return "text"
	default:
		return "ibmjson"
	}
}

func formatSimple(datetime string, message string) string {
	return fmt.Sprintf("%v %v\n", datetime, message)
}

func getDebug() bool {
	debug := os.Getenv("DEBUG")
	if debug == "true" || debug == "1" {
		return true
	}
	return false
}

func configureLogger(name string) error {
	var err error
	f := getLogFormat()
	d := getDebug()
	switch f {
	case "json":
		log, err = logger.NewLogger(os.Stderr, d, true, name)
		if err != nil {
			return err
		}
		return nil
	case "basic":
		log, err = logger.NewLogger(os.Stderr, d, false, name)
		if err != nil {
			return err
		}
		return nil
	default:
		log, err = logger.NewLogger(os.Stdout, d, false, name)
		if err != nil {
			return err
		}
		return fmt.Errorf("invalid value for LOG_FORMAT: %v", f)
	}
}
