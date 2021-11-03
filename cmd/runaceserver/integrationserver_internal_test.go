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
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ot4i/ace-docker/common/logger"
	"github.com/stretchr/testify/assert"
)

var yamlTests = []struct {
	in  string
	out string
}{
	{ // User's yaml has Statistics - it should keep accountingOrigin in Statics and any other main sections
		`Defaults:
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
{ // User's yaml has accountingOrigin in Statistics.Snapshot. It keeps this value.
`
Statistics:
 Snapshot:
  accountingOrigin: mockValue1
  nodeDataLevel: mockValue2
  outputFormat: csv
  publicationOn: mockValue3
  threadDataLevel: mockValue4`,
`Statistics:
  Resource:
    reportingOn: true
  Snapshot:
    accountingOrigin: mockValue1
    nodeDataLevel: mockValue2
    outputFormat: csv,json
    publicationOn: mockValue3
    threadDataLevel: mockValue4
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

func TestCheckLogs(t *testing.T) {
	err := checkLogs()
	if err != nil {
		t.Error(err)
	}
}


var yamlAdminTests = []struct {
	in  string
	out string
}{
	{ // User's yaml does not have a ResourceAdminListener section, so it is added
`Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'`,
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
RestAdminListener:
  caPath: /home/aceuser/adminssl
  requireClientCert: true
  sslCertificate: /home/aceuser/adminssl/tls.crt.pem
  sslPassword: /home/aceuser/adminssl/tls.key.pem
`},
	{ // User's yaml has RestAdminListener in don't alter.
`Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'
RestAdminListener:
  caPath: "test"
  requireClientCert: false
  sslCertificate: "test"
  sslPassword: "test"`,
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
RestAdminListener:
  caPath: test
  requireClientCert: false
  sslCertificate: test
  sslPassword: test
`},
	{ // User's yaml has a ResourceAdminListener section, so ours is merged with users taking precedence
`Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'
RestAdminListener:
  authorizationEnabled: true
  requireClientCert: false
  authorizationMode: file
  sslPassword: "test"
`,
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
RestAdminListener:
  authorizationEnabled: true
  authorizationMode: file
  caPath: /home/aceuser/adminssl
  requireClientCert: false
  sslCertificate: /home/aceuser/adminssl/tls.crt.pem
  sslPassword: test
`},
}


func TestAddAdminsslToServerConf(t *testing.T) {
	for _, table := range yamlAdminTests {
		out, err := addAdminsslToServerConf([]byte(table.in))
		if err != nil {
			t.Error(err)
		}
		stringOut := string(out)
		if stringOut != table.out {
			t.Errorf("addAdminsslToServerConf expected \n%v, got \n%v", table.out, stringOut)
		}
	}
}

var yamlForceFlowsHttpsTests = []struct {
	in  string
	out string
}{
	{ // User's yaml does not have a ResourceManagers section, so it is added
`Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'`,
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
ResourceManagers:
  HTTPSConnector:
    KeystoreFile: /home/aceuser/ace-server/https-keystore.p12
    KeystorePassword: brokerHTTPSKeystore::password
    KeystoreType: PKCS12
forceServerHTTPS: true
`},
	{ // User's yaml has ResourceManagers in don't alter.
`Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'
ResourceManagers:
  HTTPSConnector:
    ListenerPort: 7843
    KeystoreFile: '/home/aceuser/keystores/des-01-quickstart-ma-designer-ks'
    KeystorePassword: 'changeit'
    KeystoreType: 'PKCS12'
    CORSEnabled: true`,
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
ResourceManagers:
  HTTPSConnector:
    CORSEnabled: true
    KeystoreFile: /home/aceuser/keystores/des-01-quickstart-ma-designer-ks
    KeystorePassword: changeit
    KeystoreType: PKCS12
    ListenerPort: 7843
forceServerHTTPS: true
`},
	{ // User's yaml has a ResourceManagers HTTPSConnector section, so ours is merged with users taking precedence
`Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'
ResourceManagers:
  HTTPSConnector:
    ListenerPort: 7843
    KeystorePassword: 'changeit'
    CORSEnabled: true
`,
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
ResourceManagers:
  HTTPSConnector:
    CORSEnabled: true
    KeystoreFile: /home/aceuser/ace-server/https-keystore.p12
    KeystorePassword: changeit
    KeystoreType: PKCS12
    ListenerPort: 7843
forceServerHTTPS: true
`},
{ // User's yaml has a ResourceManagers but no HTTPSConnector section, so ours is merged with users taking precedence
`Defaults:
 defaultApplication: ''
 policyProject: 'DefaultPolicies'
 Policies:
  HTTPSConnector: 'HTTPS'
ResourceManagers:
  SOMETHINGELSE:
    ListenerPort: 9999
    KeystorePassword: 'otherbit'
    CORSEnabled: false
`,
`Defaults:
  Policies:
    HTTPSConnector: HTTPS
  defaultApplication: ""
  policyProject: DefaultPolicies
ResourceManagers:
  HTTPSConnector:
    KeystoreFile: /home/aceuser/ace-server/https-keystore.p12
    KeystorePassword: brokerHTTPSKeystore::password
    KeystoreType: PKCS12
  SOMETHINGELSE:
    CORSEnabled: false
    KeystorePassword: otherbit
    ListenerPort: 9999
forceServerHTTPS: true
`},
}

func TestAddforceFlowsHttpsToServerConf(t *testing.T) {
	for _, table := range yamlForceFlowsHttpsTests {
		out, err := addforceFlowsHttpsToServerConf([]byte(table.in))
		if err != nil {
			t.Error(err)
		}
		stringOut := string(out)
		if stringOut != table.out {
			t.Errorf("addforceFlowsHttpsToServerConf expected \n%v, got \n%v", table.out, stringOut)
		}
	}
}


func TestGetConfigurationFromContentServer(t *testing.T) {

  barDirPath := "/home/aceuser/initial-config/bars"
  var osMkdirRestore = osMkdir
  var osCreateRestore = osCreate
  var osStatRestore = osStat
  var ioutilReadFileRestore = ioutilReadFile
  var ioCopyRestore = ioCopy
  var contentserverGetBARRestore = contentserverGetBAR

  var dummyCert =  `-----BEGIN CERTIFICATE-----
  MIIC1TCCAb2gAwIBAgIJANoE+RGRB8c6MA0GCSqGSIb3DQEBBQUAMBoxGDAWBgNV
  BAMTD3d3dy5leGFtcGxlLmNvbTAeFw0yMTA1MjUwMjE3NDlaFw0zMTA1MjMwMjE3
  NDlaMBoxGDAWBgNVBAMTD3d3dy5leGFtcGxlLmNvbTCCASIwDQYJKoZIhvcNAQEB
  BQADggEPADCCAQoCggEBAOOh3VRmp/NfWbzXONxeLK/JeNvtC+TnWCz6HgtRzlhe
  7qe55dbm51Z6+l9y3C4KYH/K+a8Wgb9pKfeGCtfhRybVW3lYFtfudW7LrvgTyRIr
  r/D9UPK9J+4p/ucClGERixSY8a2F4L3Bt3o1eKDeRnz5rUlmO2mJOw41p8sSgWtp
  9MOaw5OdUrqgXh3qWkctJ8gWS2eMddh0T5ZTYE2VAOW8mTtSwAFYeBSzB+/mcl8Y
  BE7pOd71a3ka2xxLwm9KfSGLQTw0K7PxeZvdEcAq+Ffb+f/eOw0TwkNPGjnmQxLa
  MSEDDOw0AzYPibRAZIBfhLeBOHifxTd0XbCYOUAD5zkCAwEAAaMeMBwwGgYDVR0R
  BBMwEYIPd3d3LmV4YW1wbGUuY29tMA0GCSqGSIb3DQEBBQUAA4IBAQAuTsah7W7H
  HRvsuzEnPNXKKuy/aTvI3nnr6P9n5QCsdFirPyAS3/H7iNbvHVfSTfFa+b2qDaGU
  tbLJkT84m/R6gIIzRMbA0WUwQ7GRJE3KwKIytSZcTY0AuQnXy7450COmka9ztBuI
  rPYRV01LzLPJxO9A07tThSFMzhUiKrkeB5pBIjzgcYgQZNCfNtqpITmXKbO84oWA
  rbxwlF1RCvmAvzIqQx21IX16i/vH/cQ3VvCIQJt1X47KCKmWaft9AkCdjyWrFh5M
  ZhCApdQ3e/+TGkBlX32kHaRmn4Ascib7aQI2ugvowqLYFg/2LSeA0nexL+hA2GJB
  GKhFuYZvggen
  -----END CERTIFICATE-----`

  osMkdirError := errors.New("osMkdir failed")

	var reset = func() {
		osMkdir = func(path string, mode os.FileMode) (error) {
			panic("should be mocked")
		}
    osCreate = func(file string) (*os.File, error) {
			panic("should be mocked")
		}
		osStat = func(file string) (os.FileInfo, error) {
			panic("should be mocked")
		}
    ioCopy = func(target io.Writer, source io.Reader) (int64, error) {
			panic("should be mocked")
		}
    ioutilReadFile = func(cafile string) ([]byte, error) {
			panic("should be mocked")
		}
    contentserverGetBAR = func(url string, serverName string, token string, contentServerCACert []byte, contentServerCert string, contentServerKey string, log logger.LoggerInterface) (io.ReadCloser, error) {
			panic("should be mocked")
		}
	}

	var restore = func() {
		osMkdir = osMkdirRestore
    osCreate = osCreateRestore
    osStat = osStatRestore
    ioutilReadFile = ioutilReadFileRestore
    ioCopy = ioCopyRestore
    contentserverGetBAR = contentserverGetBARRestore
  }

	reset()
	defer restore()

  t.Run("No error when ACE_CONTENT_SERVER_URL not set", func(t *testing.T) {

    os.Unsetenv("ACE_CONTENT_SERVER_URL")

		osMkdir = func(dirPath string, mode os.FileMode) (error) {
      assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return nil
		}
  
    errorReturned := getConfigurationFromContentServer()

    assert.Nil(t, errorReturned)
  })

  t.Run("Fails when mkDir fails", func(t *testing.T) {

    var contentServerName = "domsdash-ibm-ace-dashboard-prod"
    var barAuth = "userid=fsdjfhksdjfhsd"
    var barUrl = "https://"+contentServerName+":3443/v1/directories/CustomerDatabaseV1"
    
    os.Unsetenv("DEFAULT_CONTENT_SERVER")
    os.Setenv("ACE_CONTENT_SERVER_URL", barUrl)
    os.Setenv("ACE_CONTENT_SERVER_NAME", contentServerName)
    os.Setenv("ACE_CONTENT_SERVER_TOKEN", barAuth)
    os.Setenv("CONTENT_SERVER_CERT", "cacert")
    os.Setenv("CONTENT_SERVER_KEY", "cakey")

		osMkdir = func(dirPath string, mode os.FileMode) (error) {
      assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return osMkdirError
		}
  
    errorReturned := getConfigurationFromContentServer()

    assert.Equal(t, errorReturned, osMkdirError)
  })

  
  t.Run("Creates barfile.bar when ACE_CONTENT_SERVER_URL is SINGLE url and  ACE_CONTENT_SERVER_NAME + ACE_CONTENT_SERVER_TOKEN  - backward compatibility for MQ connector", func(t *testing.T) {
     
    var contentServerName = "domsdash-ibm-ace-dashboard-prod"
    var barAuth = "userid=fsdjfhksdjfhsd"
    var barUrl = "https://"+contentServerName+":3443/v1/directories/CustomerDatabaseV1"
    
    testReadCloser := ioutil.NopCloser(strings.NewReader("test"))

    os.Unsetenv("DEFAULT_CONTENT_SERVER")
    os.Setenv("ACE_CONTENT_SERVER_URL", barUrl)
    os.Setenv("ACE_CONTENT_SERVER_NAME", contentServerName)
    os.Setenv("ACE_CONTENT_SERVER_TOKEN", barAuth)
    os.Setenv("CONTENT_SERVER_CERT", "cacert")
    os.Setenv("CONTENT_SERVER_KEY", "cakey")


		osMkdir = func(dirPath string, mode os.FileMode) (error) {
      assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return nil
		}
  
    osCreate = func(file string) (*os.File, error) {
      assert.Equal(t, "/home/aceuser/initial-config/bars/barfile.bar", file)
			return nil, nil
		}


		osStat = func(file string) (os.FileInfo, error) {
			// Should not be called
			t.Errorf("Should not check if file exist when only single bar URL")
			return nil, nil
		}

    ioutilReadFile = func(cafile string) ([]byte, error) {
      assert.Equal(t, "/home/aceuser/ssl/cacert.pem", cafile)
			return []byte(dummyCert), nil
		}

    contentserverGetBAR = func(url string, serverName string, token string, contentServerCACert []byte, contentServerCert string, contentServerKey string, log logger.LoggerInterface) (io.ReadCloser, error) {
      assert.Equal(t, barUrl + "?archive=true", url)
      assert.Equal(t, contentServerName, serverName)
      assert.Equal(t, barAuth, token)
      assert.Equal(t, []byte(dummyCert), contentServerCACert)
      assert.Equal(t, "cacert", contentServerCert)
      assert.Equal(t, "cakey", contentServerKey)
			return testReadCloser, nil
		}

    ioCopy = func(target io.Writer, source io.Reader) (int64, error) {
			return 0, nil
		}

    errorReturned := getConfigurationFromContentServer()

    assert.Nil(t, errorReturned)

  })

  t.Run("Error when DEFAULT_CONTENT_SERVER is true but no token found", func(t *testing.T) {
     
    var contentServerName = "domsdash-ibm-ace-dashboard-prod"
    var barUrl = "https://"+contentServerName+":3443/v1/directories/CustomerDatabaseV1"
    
    os.Setenv("DEFAULT_CONTENT_SERVER", "true")
    os.Setenv("ACE_CONTENT_SERVER_URL", barUrl)
    os.Unsetenv("ACE_CONTENT_SERVER_NAME")
    os.Unsetenv("ACE_CONTENT_SERVER_TOKEN")
    os.Setenv("CONTENT_SERVER_CERT", "cacert")
    os.Setenv("CONTENT_SERVER_KEY", "cakey")


		osMkdir = func(dirPath string, mode os.FileMode) (error) {
      assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return nil
		}
  
    ioutilReadFile = func(cafile string) ([]byte, error) {
      assert.Equal(t, "/home/aceuser/ssl/cacert.pem", cafile)
			return []byte(dummyCert), nil
		}

    errorReturned := getConfigurationFromContentServer()

		assert.Equal(t, errors.New("No content server token available but a url is defined"), errorReturned)

  })

  t.Run("No error when DEFAULT_CONTENT_SERVER is false and token found", func(t *testing.T) {
     
    var contentServerName = "domsdash-ibm-ace-dashboard-prod"
    var barUrl = "https://"+contentServerName+":3443/v1/directories/CustomerDatabaseV1?user=default"
    
    testReadCloser := ioutil.NopCloser(strings.NewReader("test"))

    os.Setenv("DEFAULT_CONTENT_SERVER", "false")
    os.Setenv("ACE_CONTENT_SERVER_URL", barUrl)
    os.Unsetenv("ACE_CONTENT_SERVER_NAME")
    os.Unsetenv("ACE_CONTENT_SERVER_TOKEN")
    os.Setenv("CONTENT_SERVER_CA", "/home/aceuser/ssl/mycustom.pem")
    os.Setenv("CONTENT_SERVER_CERT", "cacert")
    os.Setenv("CONTENT_SERVER_KEY", "cakey")


		osMkdir = func(dirPath string, mode os.FileMode) (error) {
      assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return nil
		}
  
    osCreate = func(file string) (*os.File, error) {
      assert.Equal(t, "/home/aceuser/initial-config/bars/barfile.bar", file)
			return nil, nil
		}

		osStat = func(file string) (os.FileInfo, error) {
			// Should not be called
			t.Errorf("Should not check if file exist when only single bar URL")
			return nil, nil
		}

    ioutilReadFile = func(cafile string) ([]byte, error) {
      assert.Equal(t, "/home/aceuser/ssl/mycustom.pem", cafile)
			return []byte(dummyCert), nil
		}

    contentserverGetBAR = func(url string, serverName string, token string, contentServerCACert []byte, contentServerCert string, contentServerKey string, log logger.LoggerInterface) (io.ReadCloser, error) {
      assert.Equal(t, barUrl, url)
      assert.Equal(t, contentServerName, serverName)
      assert.Equal(t, "user=default", token)
      assert.Equal(t, []byte(dummyCert), contentServerCACert)
      assert.Equal(t, "cacert", contentServerCert)
      assert.Equal(t, "cakey", contentServerKey)
			return testReadCloser, nil
		}

    ioCopy = func(target io.Writer, source io.Reader) (int64, error) {
			return 0, nil
		}

    errorReturned := getConfigurationFromContentServer()

    assert.Nil(t, errorReturned)

  })


  t.Run("No error when DEFAULT_CONTENT_SERVER is false and no token found", func(t *testing.T) {
     
    var contentServerName = "domsdash-ibm-ace-dashboard-prod"
    var barUrl = "https://"+contentServerName+":3443/v1/directories/CustomerDatabaseV1"
    
    testReadCloser := ioutil.NopCloser(strings.NewReader("test"))

    os.Setenv("DEFAULT_CONTENT_SERVER", "false")
    os.Setenv("ACE_CONTENT_SERVER_URL", barUrl)
    os.Unsetenv("ACE_CONTENT_SERVER_NAME")
    os.Unsetenv("ACE_CONTENT_SERVER_TOKEN")
    os.Setenv("CONTENT_SERVER_CA", "/home/aceuser/ssl/mycustom.pem")
    os.Setenv("CONTENT_SERVER_CERT", "cacert")
    os.Setenv("CONTENT_SERVER_KEY", "cakey")


		osMkdir = func(dirPath string, mode os.FileMode) (error) {
      assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return nil
		}
  
    osCreate = func(file string) (*os.File, error) {
      assert.Equal(t, "/home/aceuser/initial-config/bars/barfile.bar", file)
			return nil, nil
		}

		osStat = func(file string) (os.FileInfo, error) {
			// Should not be called
			t.Errorf("Should not check if file exist when only single bar URL")
			return nil, nil
		}

    ioutilReadFile = func(cafile string) ([]byte, error) {
      assert.Equal(t, "/home/aceuser/ssl/mycustom.pem", cafile)
			return []byte(dummyCert), nil
		}

    contentserverGetBAR = func(url string, serverName string, token string, contentServerCACert []byte, contentServerCert string, contentServerKey string, log logger.LoggerInterface) (io.ReadCloser, error) {
      assert.Equal(t, barUrl, url)
      assert.Equal(t, contentServerName, serverName)
      assert.Equal(t, "", token)
      assert.Equal(t, []byte(dummyCert), contentServerCACert)
      assert.Equal(t, "cacert", contentServerCert)
      assert.Equal(t, "cakey", contentServerKey)
			return testReadCloser, nil
		}

    ioCopy = func(target io.Writer, source io.Reader) (int64, error) {
			return 0, nil
		}

    errorReturned := getConfigurationFromContentServer()

    assert.Nil(t, errorReturned)

  })


  t.Run("Error when DEFAULT_CONTENT_SERVER is false but no CONTENT_SERVER_CA found", func(t *testing.T) {
     
    var contentServerName = "domsdash-ibm-ace-dashboard-prod"
    var barUrl = "https://"+contentServerName+":3443/v1/directories/CustomerDatabaseV1"
    
    os.Setenv("DEFAULT_CONTENT_SERVER", "false")
    os.Setenv("ACE_CONTENT_SERVER_URL", barUrl)
    os.Unsetenv("ACE_CONTENT_SERVER_NAME")
    os.Unsetenv("ACE_CONTENT_SERVER_TOKEN")
    os.Unsetenv("CONTENT_SERVER_CA")
    os.Setenv("CONTENT_SERVER_CERT", "cacert")
    os.Setenv("CONTENT_SERVER_KEY", "cakey")


		osMkdir = func(dirPath string, mode os.FileMode) (error) {
      assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return nil
		}

    errorReturned := getConfigurationFromContentServer()

		assert.Equal(t, errors.New("CONTENT_SERVER_CA not defined"), errorReturned)

  })
	
  t.Run("Creates multiple files when ACE_CONTENT_SERVER_URL is array url and extracts server name and auth from urls  - multi bar support", func(t *testing.T) {
    
    //https://alexdash-ibm-ace-dashboard-prod:3443/v1/directories/CustomerDatabaseV1?userid=fsdjfhksdjfhsd
    var barName1 = "CustomerDatabaseV1"
    var contentServerName1 = "alexdash-ibm-ace-dashboard-prod"
    var barAuth1 = "userid=fsdjfhksdjfhsd"
    var barUrl1 = "https://"+contentServerName1+":3443/v1/directories/" + barName1 
    
    //https://test-acecontentserver-ace-alex.svc:3443/v1/directories/testdir?e31d23f6-e3ba-467d-ab3b-ceb0ab12eead
    var barName2 = "testdir"
    var contentServerName2 = "test-acecontentserver-ace-alex.svc"
    var barAuth2 = "e31d23f6-e3ba-467d-ab3b-ceb0ab12eead"
    var barUrl2 = "https://"+contentServerName2+":3443/v1/directories/" + barName2

    var barUrl = barUrl1 + "?" + barAuth1 + "," + barUrl2 + "?" + barAuth2

    testReadCloser := ioutil.NopCloser(strings.NewReader("test"))

    os.Unsetenv("DEFAULT_CONTENT_SERVER")
    os.Setenv("ACE_CONTENT_SERVER_URL", barUrl)
    os.Unsetenv("ACE_CONTENT_SERVER_NAME")
    os.Unsetenv("ACE_CONTENT_SERVER_TOKEN")
    os.Setenv("CONTENT_SERVER_CERT", "cacert")
    os.Setenv("CONTENT_SERVER_KEY", "cakey")

		osMkdir = func(dirPath string, mode os.FileMode) (error) {
      assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return nil
		}
  
    osCreateCall := 1
    osCreate = func(file string) (*os.File, error) {
      if osCreateCall == 1 {
        assert.Equal(t, "/home/aceuser/initial-config/bars/" + barName1 + ".bar", file)
      } else if osCreateCall == 2 {
        assert.Equal(t, "/home/aceuser/initial-config/bars/" + barName2 + ".bar", file)
      }
      osCreateCall = osCreateCall + 1
			return nil, nil
		}

    osStat = func(file string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}

    ioutilReadFile = func(cafile string) ([]byte, error) {
      assert.Equal(t, "/home/aceuser/ssl/cacert.pem", cafile)
			return []byte(dummyCert), nil
		}

    getBarCall := 1
    contentserverGetBAR = func(url string, serverName string, token string, contentServerCACert []byte, contentServerCert string, contentServerKey string, log logger.LoggerInterface) (io.ReadCloser, error) {
      if getBarCall == 1 {
        assert.Equal(t, barUrl1 + "?archive=true", url)
        assert.Equal(t, contentServerName1, serverName)
        assert.Equal(t, barAuth1, token)
        assert.Equal(t, []byte(dummyCert), contentServerCACert)
        assert.Equal(t, "cacert", contentServerCert)
        assert.Equal(t, "cakey", contentServerKey)
      } else if getBarCall == 2 {
        assert.Equal(t, barUrl2 + "?archive=true", url)
        assert.Equal(t, contentServerName2, serverName)
        assert.Equal(t, barAuth2, token)
        assert.Equal(t, []byte(dummyCert), contentServerCACert)
        assert.Equal(t, "cacert", contentServerCert)
        assert.Equal(t, "cakey", contentServerKey)
      }
      getBarCall = getBarCall + 1
      return testReadCloser, nil
		}

    ioCopy = func(target io.Writer, source io.Reader) (int64, error) {
			return 0, nil
		}

    errorReturned := getConfigurationFromContentServer()

    assert.Nil(t, errorReturned)
  })

	t.Run("Creates multiple files with different names when using multi bar support and the bar file names are all the same", func(t *testing.T) {

		//https://alexdash-ibm-ace-dashboard-prod:3443/v1/directories/CustomerDatabaseV1?userid=fsdjfhksdjfhsd
		var barName = "CustomerDatabaseV1"
		var contentServerName = "alexdash-ibm-ace-dashboard-prod"
		var barAuth = "userid=fsdjfhksdjfhsd"
		var barUrlBase = "https://" + contentServerName + ":3443/v1/directories/" + barName
		var barUrlFull = barUrlBase + "?" + barAuth

		var barUrl = barUrlFull + "," + barUrlFull + "," + barUrlFull

		testReadCloser := ioutil.NopCloser(strings.NewReader("test"))

		os.Unsetenv("DEFAULT_CONTENT_SERVER")
		os.Setenv("ACE_CONTENT_SERVER_URL", barUrl)
		os.Unsetenv("ACE_CONTENT_SERVER_NAME")
		os.Unsetenv("ACE_CONTENT_SERVER_TOKEN")
		os.Setenv("CONTENT_SERVER_CERT", "cacert")
		os.Setenv("CONTENT_SERVER_KEY", "cakey")

		osMkdir = func(dirPath string, mode os.FileMode) error {
			assert.Equal(t, barDirPath, dirPath)
			assert.Equal(t, os.ModePerm, mode)
			return nil
		}

		createdFiles := map[string]bool{}
		osCreateCall := 1
		osCreate = func(file string) (*os.File, error) {
			createdFiles[file] = true
			if osCreateCall == 1 {
				assert.Equal(t, "/home/aceuser/initial-config/bars/"+barName+".bar", file)
			} else if osCreateCall == 2 {
				assert.Equal(t, "/home/aceuser/initial-config/bars/"+barName+"-1.bar", file)
			} else if osCreateCall == 3 {
				assert.Equal(t, "/home/aceuser/initial-config/bars/"+barName+"-2.bar", file)
			}
			osCreateCall = osCreateCall + 1
			return nil, nil
		}

		osStat = func(file string) (os.FileInfo, error) {
			if createdFiles[file] {
				return nil, os.ErrExist
			} else {
				return nil, os.ErrNotExist
			}
		}

		ioutilReadFile = func(cafile string) ([]byte, error) {
			assert.Equal(t, "/home/aceuser/ssl/cacert.pem", cafile)
			return []byte(dummyCert), nil
		}

		getBarCall := 1
		contentserverGetBAR = func(url string, serverName string, token string, contentServerCACert []byte, contentServerCert string, contentServerKey string, log logger.LoggerInterface) (io.ReadCloser, error) {
			assert.Equal(t, barUrlBase+"?archive=true", url)
			assert.Equal(t, contentServerName, serverName)
			assert.Equal(t, barAuth, token)
			assert.Equal(t, []byte(dummyCert), contentServerCACert)
			assert.Equal(t, "cacert", contentServerCert)
			assert.Equal(t, "cakey", contentServerKey)
			getBarCall = getBarCall + 1
			return testReadCloser, nil
		}

		ioCopy = func(target io.Writer, source io.Reader) (int64, error) {
			return 0, nil
		}

		errorReturned := getConfigurationFromContentServer()

		assert.Nil(t, errorReturned)
	})
}

func TestWatchForceFlowsHTTPSSecret(t *testing.T) {

  patchHTTPSConnectorCalled := 0
  oldpatchHTTPSConnector := patchHTTPSConnector
  defer func() { patchHTTPSConnector = oldpatchHTTPSConnector }()
  patchHTTPSConnector = func(string) {
    patchHTTPSConnectorCalled++
  }

  oldgenerateHTTPSKeystore := generateHTTPSKeystore
  defer func() { generateHTTPSKeystore = oldgenerateHTTPSKeystore }()
  generateHTTPSKeystore = func(string,  string,  string,  string) {
  }

  // Tidy up any /tmp files
  files, err := filepath.Glob("/tmp/forceflowtest*")
  if err != nil {
    log.Errorf("Error finding tmp files: %v", err)
  }
  for _, f := range files {
    if err := os.Remove(f); err != nil {
      log.Errorf("Error removing tmp files: %v", err)
    }
  }

  // Create new test file
  extension := generatePassword(5)
  f, err := os.Create("/tmp/forceflowtest"+ extension)
  helloFile := []byte{115, 111, 109, 101, 10} // hello
  _, err = f.Write(helloFile)

  // Now create the watcher and watch the file created above
  watcher := watchForceFlowsHTTPSSecret("TESTPASS")
	err = watcher.Add("/tmp/forceflowtest"+ extension)
	if err != nil {
		log.Errorf("Error watching /home/aceuser/httpsNodeCerts for Force Flows to be HTTPS: %v", err)
	}

  // Now write to the file to check we are called 2 times
  // waits are required as if you update too quickly you only get called once
  _, err = f.Write(helloFile)
  time.Sleep(2* time.Second)
  _, err = f.Write(helloFile)
  time.Sleep(2* time.Second)

  // clean up watcher and temporary file
  os.Remove("/tmp/forceflowtest"+ extension)
  watcher.Close()

  // expected patchHTTPSConnector to be called twice
  assert.Equal(t, patchHTTPSConnectorCalled, 2)

  log.Println("done")
}

func TestUDSCall(t *testing.T) {
  // curl --unix-socket /tmp/47EBPflowtest.uds http:/apiv2/resource-managers/https-connector/refresh-tls-config

  // Tidy up any /tmp files
  files, err := filepath.Glob("/tmp/*flowtest.uds")
  if err != nil {
    log.Errorf("Error finding tmp files: %v", err)
  }
  for _, f := range files {
    if err := os.Remove(f); err != nil {
      log.Errorf("Error removing tmp files: %v", err)
    }
  }

  extension := generatePassword(5)
  uds := "/tmp/"+extension+"flowtest.uds"

  timesServerCalled := 0

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if (r.Method == "POST") && (r.RequestURI== "/apiv2/resource-managers/https-connector/refresh-tls-config") {
      timesServerCalled++
    } else {
      log.Print("Error: TestUDSCall called with the following but expected POST and /apiv2/resource-managers/https-connector/refresh-tls-config")
      log.Printf(r.Method)
      log.Printf(r.RequestURI)
      assert.Equal(t, r.Method, "POST")
      assert.Equal(t, r.RequestURI, "/apiv2/resource-managers/https-connector/refresh-tls-config")
    }
  })

  unixListener, err := net.Listen("unix", uds)
  if err != nil {
    log.Errorf("Error creating UDS listener: %v", err)
  }
  handler := http.DefaultServeMux
  // Kick off server in the background
  go func() {
    err = http.Serve(unixListener, handler)
    if err != nil {
      log.Errorf("Error starting up http UDS server: %v", err)
    }
  }()

  // test we can call the server above twice
  localPatchHTTPSConnector(uds)
  assert.Equal(t, timesServerCalled, 1)
  localPatchHTTPSConnector(uds)
  assert.Equal(t, timesServerCalled, 2)

  // tidy up uds file
  os.Remove(uds)
}



func TestGenerateHTTPSKeystore(t *testing.T) {
  password := "anypassword"

  testPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`

  testCertificate := `-----BEGIN CERTIFICATE-----
MIIF0DCCA7igAwIBAgIQDqKhyQoLfI24yKy9pB9CajANBgkqhkiG9w0BAQsFADCB
ijELMAkGA1UEBhMCR0IxEjAQBgNVBAgTCUhhbXBzaGlyZTETMBEGA1UEBxMKV2lu
Y2hlc3RlcjEjMCEGA1UEChMaSUJNIFVuaXRlZCBLaW5nZG9tIExpbWl0ZWQxFDAS
BgNVBAsTC0FwcCBDb25uZWN0MRcwFQYDVQQDEw5hcHAtY29ubmVjdC1jYTAeFw0y
MTA3MDgwOTQ3NTZaFw0zMTA3MDgwOTQ3NTZaMIGMMQswCQYDVQQGEwJHQjESMBAG
A1UECBMJSGFtcHNoaXJlMRMwEQYDVQQHEwpXaW5jaGVzdGVyMSMwIQYDVQQKExpJ
Qk0gVW5pdGVkIEtpbmdkb20gTGltaXRlZDEUMBIGA1UECxMLQXBwIENvbm5lY3Qx
GTAXBgNVBAMTEGlzLTAxLXRvb2xraXQtaXMwggIiMA0GCSqGSIb3DQEBAQUAA4IC
DwAwggIKAoICAQDrdSVU4RaQckUs8j88AdmpljKXj+cT6kMVcpdi5SNFGNc/2tgA
QrfOfwHnKw/BOsGr9EjuMIpPKIf5CuIgW8g6O56tD4Vg2T3wXBDDg1sx6CHOKKL7
0TSq13xVDdBL9J0LYXkWi3SpUjOgZSTaMfr8i+m9PpQZV13XiklGQ6kYUuuiNwjw
zp303xGQxZMi+bFuHDkemWxoxXe9mENXjHyAqD9s9J3d13yKfOB6cSeNLprbAJtq
5nIg1QI4oCgunOyTMLCk2zWIRoP7llOqwRsMMsRTXTQWubjHbWQgY6YA179iqtKD
k/ZSH+MsWht5hVmd8QLtymSxtL3gA9MDn7hMM2Pum/DSBxAZNm2L//A7VuKQoV9k
quoxG+YVuqp6CgrSVCMxcU59K5wjpyhwQHachxk05ivlbXsYQZ2sfSWYC//Uv2sM
HLlpo1w1+pTsgcQN6iq3txxypd9lXqlV5ALH794EoTUP4eYngywd3ULpe8AmsEq2
LAJ7k+M6zmKJniyRJ+2NVOvxBX7y6ev+w4nZRBtrtro4ziBbgyAR4Fm/dNv/ZpL9
ENmB8wjSQ9NKxpG90nNm1h/9cBUfrgbrQJsZ4bF+j4ypERRKzIh07oE5ygBFhKHY
TWMoeWKRdxGuRWpb+W6ti6XHelV/WLID0dBuldHzVUpIGUaFBicKTrYnSwIDAQAB
oy4wLDAqBgNVHREEIzAhgg1pcy0wMS10b29sa2l0ghBpcy0wMS10b29sa2l0LWlz
MA0GCSqGSIb3DQEBCwUAA4ICAQCxCR8weIn4IfHEZFUKFFAyijvq0rapWVFHaRoo
M0LPnbllyjX9H4oiDnAlwnlB4gWM06TltlJ5fR+O5Lf2aRPlvxm5mAZHKMSlwenO
DgnWSfXu5OwbnHrBVim+zcf4wmYo89HpH2UsbpVty28UjZ6elJzwkYG1MWWvWiLR
U28vps50UxuG1DQyMRiylnTKIWUdRAZG4k855UIIS9c++iCNY9S9DHAFj1Bl4hG8
N1Jfsy8IJ+wAm/QPdE1cJO4U9ky7cB32IDynuv4nBr/K3XPXu0qvPVKl9jqGogpX
FspClOHor+7c47vusvA/Cvkrn14+BRgR/HcxrEZp15Tx9Vhr6sYTpXrLbOzTGoty
KI9eRbiXt50l347ZFhvgHBuSYW8YXzZ+pFymbh2LThC6Oum167Sb5IftAH1uQvb0
WZKoaNc0JcpVkDmHkHhjLDU1G6rI/T9S7Rk/yGweLmQOVccw5y/E92KKeZzeVrNW
/LH1e0LeQkd+8KhaSWqAjxPFBHFuPF1fQVDs4OfzqgwDI40prvRQzAhkFy3TIMHK
Pr6B0s5nfURsB2sT7PUWYijTHvuuuyb5F/OLNIRXhMfGKfTwMHnhTanmZnjqQz47
Y0UDx8nGWNT0OZrP0h/IFVUNCK7oupx0S0QxiRqmGS/4TfKZR6F/Pv5Vtc9KDxvf
iipaBg==
-----END CERTIFICATE-----
`

  // Tidy up any /tmp files
  files, err := filepath.Glob("/tmp/*generatekeystore*")
  if err != nil {
    log.Errorf("Error finding tmp files: %v", err)
  }
  for _, f := range files {
    if err := os.Remove(f); err != nil {
      log.Errorf("Error removing tmp files: %v", err)
    }
  }

  extension := generatePassword(5)
  privateKeyFileLocation := "/tmp/"+extension+"generatekeystore.tls.key"
  certificateLocation := "/tmp/"+extension+"generatekeystore.tls.cert"
  keystoreLocation := "/tmp/"+extension+"generatekeystore.https-keystore.p12"

  keyFile, err := os.Create(privateKeyFileLocation)
  if (err != nil) {
    log.Error("Could not create " + privateKeyFileLocation)
    assert.Nil(t, err)
  }
  keyFile.WriteString(testPrivateKey)

  certificateFile, err := os.Create(certificateLocation)
  if (err != nil) {
    log.Error("Could not create " + certificateLocation)
    assert.Nil(t, err)
  }
  certificateFile.WriteString(testCertificate)

  generateHTTPSKeystore(privateKeyFileLocation, certificateLocation, keystoreLocation, password)

  // Check created keystore
  fileInfo, err := os.Stat(keystoreLocation)
  if (err != nil) {
    log.Error("Could not stat " + keystoreLocation)
    assert.Nil(t, err)
  }

  // Check keystore size
  assert.Equal(t, fileInfo.Size(), int64(4239))
  assert.Equal(t, fileInfo.IsDir(), false)

  // clean up
  os.Remove(privateKeyFileLocation)
  os.Remove(certificateLocation)
  os.Remove(keystoreLocation)

}