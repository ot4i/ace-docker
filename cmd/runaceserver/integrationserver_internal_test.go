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
  "strings"
	"io"
  "io/ioutil"
	"os"
	"testing"
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
