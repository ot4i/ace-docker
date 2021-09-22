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

package logger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ot4i/ace-docker/common/logger"
)

func TestJSONLogger(t *testing.T) {
	const logSourceCRN, saveServiceCopy, newParam = "testCRN", false, "123"
	os.Setenv("MQSI_LOG_ADDITIONAL_JSON_ELEMENTS", fmt.Sprintf("\"logSourceCRN\":%q, \"saveServiceCopy\":%t, \"newParam\":%q", logSourceCRN, saveServiceCopy, newParam))

	buf := new(bytes.Buffer)
	l, err := logger.NewLogger(buf, true, true, t.Name())
	if err != nil {
		t.Fatal(err)
	}
	s := "Hello world"
	l.Print(s)
	var e map[string]interface{}
	err = json.Unmarshal([]byte(buf.String()), &e)
	if err != nil {
		t.Error(err)
	}
	if s != e["message"] {
		t.Errorf("Expected JSON to contain message=%v; got %v", s, buf.String())
	}

	if e["logSourceCRN"] != logSourceCRN {
		t.Errorf("Expected JSON to contain logSourceCRN=%v; got %v", e["logSourceCRN"], buf.String())
	}

	if e["saveServiceCopy"] != saveServiceCopy {
		t.Errorf("Expected JSON to contain saveServiceCopy=%v; got %v", e["saveServiceCopy"], buf.String())
	}

	if e["newParam"] != newParam {
		t.Errorf("Expected JSON to contain newParam=%v; got %v", e["newParam"], buf.String())
	}
}

func TestSimpleLogger(t *testing.T) {
	os.Setenv("MQSI_LOG_ADDITIONAL_JSON_ELEMENTS", "\"logSourceCRN\":\"testCRN\"")

	buf := new(bytes.Buffer)
	l, err := logger.NewLogger(buf, true, false, t.Name())
	if err != nil {
		t.Fatal(err)
	}
	s := "Hello world"
	l.Print(s)
	if !strings.Contains(buf.String(), s) {
		t.Errorf("Expected log output to contain %v; got %v", s, buf.String())
	}

	if strings.Contains(buf.String(), "logSourceCRN") {
		t.Errorf("Expected log output to without %v; got %v", "logSourceCRN", buf.String())
	}
}
