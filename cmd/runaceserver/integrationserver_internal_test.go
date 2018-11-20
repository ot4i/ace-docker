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
	"testing"
)

var yamlTests = []struct {
	in  string
	out string
}{
{ // User's yaml has Statistics - it should keep accountingOrigin in Statics and any other main sections
`
Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'
Statistics:
 Snapshot:
  accountingOrigin: 'test'
 Archive:
  test1: 'blah'
  test2: 'blah2'`, 
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
Statistics:
  Archive:
    test1: blah
    test2: blah2
  Resource:
    reportingOn: true
  Snapshot:
    accountingOrigin: test
    nodeDataLevel: basic
    outputFormat: json
    publicationOn: active
    threadDataLevel: none
`},
{ // User's yaml does not have a Statistics section, it adds the default metrics info
`
Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'`, 
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
Statistics:
  Resource:
    reportingOn: true
  Snapshot:
    accountingOrigin: none
    nodeDataLevel: basic
    outputFormat: json
    publicationOn: active
    threadDataLevel: none
`},
{ // User's yaml has accountingOrigin in Statistics.Snapshot. It keeps this value.
`
Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'
Statistics:
 Snapshot:
  accountingOrigin: 'test'`, 
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
Statistics:
  Resource:
    reportingOn: true
  Snapshot:
    accountingOrigin: test
    nodeDataLevel: basic
    outputFormat: json
    publicationOn: active
    threadDataLevel: none
`},
}

func TestAddMetricsToServerConf(t *testing.T) {
	for _, table := range yamlTests {
		out, err := addMetricsToServerConf([]byte(table.in))
		if err != nil {
			t.Error(err)
		}
		stringOut := string(out)
		if stringOut != table.out {
			t.Errorf("addMetricsToServerConf expected %v, got %v", table.out, stringOut)
		}
	}
}
