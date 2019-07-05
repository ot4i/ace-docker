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
package qmgr_test

import (
	"os"
	"testing"

	"github.com/ot4i/ace-docker/internal/qmgr"
)

var useTests = []struct {
	in  string
	out bool
}{
	{"true", true},
	{"yes", false},
	{"false", false},
}

func TestUseQueueManager(t *testing.T) {
	for _, table := range useTests {

		err := os.Setenv("USE_QMGR", table.in)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		b := qmgr.UseQueueManager()
		if b != table.out {
			t.Errorf("qmgr.UseQueueManager() with USE_QMGR=%v - expected %v, got %v", table.in, table.out, b)
		}
	}
}

func TestDevManager(t *testing.T) {
	for _, table := range useTests {

		err := os.Setenv("DEV_QMGR", table.in)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		b := qmgr.DevManager()
		if b != table.out {
			t.Errorf("qmgr.DevManager() with DEV_QMGR=%v - expected %v, got %v", table.in, table.out, b)
		}
	}
}
