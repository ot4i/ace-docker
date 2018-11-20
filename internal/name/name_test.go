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
package name_test

import (
	"os"
	"testing"

	"github.com/ot4i/ace-docker/internal/name"
)

func TestGetQueueManagerNameFromEnv(t *testing.T) {
	const data string = "foo"
	err := os.Setenv("MQ_QMGR_NAME", data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	n, err := name.GetQueueManagerName()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != data {
		t.Errorf("Expected name=%v, got name=%v", data, n)
	}
}

func TestGetIntegrationServerNameFromEnv(t *testing.T) {
	const data string = "bar"
	err := os.Setenv("ACE_SERVER_NAME", data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	n, err := name.GetIntegrationServerName()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != data {
		t.Errorf("Expected name=%v, got name=%v", data, n)
	}
}
