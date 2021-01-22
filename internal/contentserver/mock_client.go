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
	"net/http"
)

type Client interface {
	Do(*http.Request) (*http.Response, error) 
}

type MockClient struct {
	Client *http.Client
}

var response *http.Response = nil
var err error = nil
func (m *MockClient) Do(request *http.Request) (*http.Response, error) {
	if response != nil || err != nil{
		return response, err
	}
	return m.Client.Do(request)
}

func setDoResponse(mockResponse *http.Response, mockError error) {
	response = mockResponse
	err = mockError
}

func resetDoResponse() {
	response = nil
	err = nil
}