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
package contentserver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ot4i/ace-docker/common/logger"
)

var loadX509KeyPair = tls.LoadX509KeyPair
var newRequest = http.NewRequest

var newHTTPClient = func(rootCAs *x509.CertPool, cert tls.Certificate, serverName string) Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      rootCAs,
				Certificates: []tls.Certificate{cert},
				ServerName:   serverName,
			},
		},
	}
}

func GetBAR(url string, serverName string, token string, contentServerCACert []byte, contentServerCert string, contentServerKey string, log logger.LoggerInterface) (io.ReadCloser, error) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(contentServerCACert)

	// If provided read the key pair to create certificate
	cert, err := loadX509KeyPair(contentServerCert, contentServerKey)
	if err != nil {
		if contentServerCert != "" && contentServerKey != "" {
			log.Errorf("Error reading Certificates: %s", err)
			return nil, errors.New("Error reading Certificates")
		}
	} else {
		log.Printf("Using certs for mutual auth")
	}

	client := newHTTPClient(caCertPool, cert, serverName)

	request, err := newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("x-ibm-ace-directory-token", token)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		log.Printf("Call to retrieve BAR file from content server failed with response code: %v", response.StatusCode)
		return nil, errors.New("HTTP status : " + fmt.Sprint(response.StatusCode) + ". Unable to get BAR file from content server.")
	}

	return response.Body, nil
}
