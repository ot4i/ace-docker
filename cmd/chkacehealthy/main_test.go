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

// chkacelively checks that ACE is still runing, by checking if the admin REST endpoint port is available.
package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
)

func Test_httpChecklocal(t *testing.T) {

	t.Run("http get succeeds", func(t *testing.T) {

		oldhttpGet := httpGet
		defer func() { httpGet = oldhttpGet }()
		httpGet = func(string) (resp *http.Response, err error) {
			response := &http.Response{
				StatusCode: 200,
			}
			return response, nil
		}

		err := httpChecklocal("LMAP Port", "http://localhost:3002/admin/ready")
		assert.Nil(t, err)
	})

	t.Run("http get fails with err on get", func(t *testing.T) {

		oldhttpGet := httpGet
		defer func() { httpGet = oldhttpGet }()
		httpGet = func(string) (resp *http.Response, err error) {
			response := &http.Response{}
			return response, errors.NewBadRequest("mock err")
		}

		err := httpChecklocal("LMAP Port", "http://localhost:3002/admin/ready")
		assert.Error(t, err, "mock err")
	})

	t.Run("http get fails with non 200", func(t *testing.T) {

		oldhttpGet := httpGet
		defer func() { httpGet = oldhttpGet }()
		httpGet = func(string) (resp *http.Response, err error) {
			response := &http.Response{
				StatusCode: 404,
			}
			return response, nil
		}

		err := httpChecklocal("Test", "http://localhost:3002/admin/ready")
		assert.Error(t, err, "Test ready check failed - HTTP Status is not 200 range")
	})
}

func Test_checkDesignerHealth(t *testing.T) {
	t.Run("connector service enabled and health check is successful", func(t *testing.T) {
		os.Setenv("CONNECTOR_SERVICE", "true")
		defer os.Unsetenv("CONNECTOR_SERVICE")
		oldhttpGet := httpGet
		defer func() { httpGet = oldhttpGet }()
		httpGet = func(string) (resp *http.Response, err error) {
			response := &http.Response{
				StatusCode: 200,
			}
			return response, nil
		}

		oldSocketCheck := socketCheck
		defer func() { socketCheck = oldSocketCheck }()
		socketCheck = func(string, string) (err error) {
			return nil
		}
		err := checkDesignerHealth()
		assert.Nil(t, err)
	})

	t.Run("connector service enabled and health check fails", func(t *testing.T) {
		os.Setenv("CONNECTOR_SERVICE", "true")
		defer os.Unsetenv("CONNECTOR_SERVICE")
		oldhttpGet := httpGet
		defer func() { httpGet = oldhttpGet }()
		httpGet = func(string) (resp *http.Response, err error) {
			response := &http.Response{}
			return response, errors.NewBadRequest("mock err")
		}

		oldSocketCheck := socketCheck
		defer func() { socketCheck = oldSocketCheck }()
		socketCheck = func(string, string) (err error) {
			return nil
		}
		err := checkDesignerHealth()
		assert.Error(t, err, "mock err")
	})

	t.Run("health check fails for socket http server", func(t *testing.T) {
		oldhttpGet := httpGet
		defer func() { httpGet = oldhttpGet }()
		httpGet = func(string) (resp *http.Response, err error) {
			response := &http.Response{
				StatusCode: 200,
			}
			return response, nil
		}

		oldSocketCheck := socketCheck
		defer func() { socketCheck = oldSocketCheck }()
		socketCheck = func(string, string) (err error) {
			return errors.NewBadRequest("mock err")
		}
		err := checkDesignerHealth()
		assert.Error(t, err, "mock err")
	})
}
