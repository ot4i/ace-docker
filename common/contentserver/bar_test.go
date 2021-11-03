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

package contentserver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ot4i/ace-docker/common/logger"

	"testing"

	"github.com/stretchr/testify/assert"
)

var testLogger, _ = logger.NewLogger(os.Stdout, true, true, "test")

func TestGetBAR(t *testing.T) {
	url := ""
	serverName := ""
	token := ""
	contentServerCACert := []byte{}

	oldNewHTTPClient := newHTTPClient
	newHTTPClient = func(*x509.CertPool, tls.Certificate, string) Client {
		return &MockClient{Client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}}
	}

	t.Run("When not able to load the content server cert and key pair, it returns an error", func(t *testing.T) {
		oldLoadX509KeyPair := loadX509KeyPair
		loadX509KeyPair = func(string, string) (tls.Certificate, error) {
			return tls.Certificate{}, errors.New("Fail to load key pair")
		}

		t.Run("When the server cert and key are not empty strings, it fails", func(t *testing.T) {
			_, err := GetBAR(url, serverName, token, contentServerCACert, "contentServerCert", "contentServerKey", testLogger)
			assert.Error(t, err)
		})

		loadX509KeyPair = oldLoadX509KeyPair
	})

	t.Run("When failing to create the http requestv", func(t *testing.T) {
		oldNewRequest := newRequest
		newRequest = func(string, string, io.Reader) (*http.Request, error) {
			return nil, errors.New("Fail to create new request")
		}

		_, err := GetBAR(url, serverName, token, contentServerCACert, "", "", testLogger)
		assert.Error(t, err)

		newRequest = oldNewRequest
	})

	t.Run("When failing to make the client call, it returns an error", func(t *testing.T) {
		setDoResponse(nil, errors.New("Fail to create new request"))
		_, err := GetBAR(url, serverName, token, contentServerCACert, "", "", testLogger)
		assert.Error(t, err)
		resetDoResponse()
	})

	// TODO: should this return an error?
	t.Run("When the client call reponds with a non 200, it does return an error", func(t *testing.T) {
		setDoResponse(&http.Response{StatusCode: 500}, nil)
		_, err := GetBAR(url, serverName, token, contentServerCACert, "", "", testLogger)
		assert.Error(t, err)
		resetDoResponse()
	})

	t.Run("When the client call reponds with a 200, it returns the body", func(t *testing.T) {
		testReadCloser := ioutil.NopCloser(strings.NewReader("test"))
		setDoResponse(&http.Response{StatusCode: 200, Body: testReadCloser}, nil)
		body, err := GetBAR(url, serverName, token, contentServerCACert, "", "", testLogger)
		assert.NoError(t, err)
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)
		assert.Equal(t, "test", buf.String())
		resetDoResponse()
	})
	newHTTPClient = oldNewHTTPClient
}
