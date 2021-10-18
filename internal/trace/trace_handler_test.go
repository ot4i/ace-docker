/*
© Copyright IBM Corporation 2021

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

package trace

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ot4i/ace-docker/common/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testLogger, _ = logger.NewLogger(os.Stdout, true, true, "test")

func TestStartTraceServer(t *testing.T) {
	log = testLogger

	t.Run("Error creating cert pool", func(t *testing.T) {
		tlsEnabled = true
		caPath = "capath"
		err := startTraceServer(&Server{}, 7891)
		require.Error(t, err)
	})

	t.Run("TLS server when ACE_ADMIN_SERVER_SECURITY is true", func(t *testing.T) {
		defer func() {
			os.RemoveAll("capath")
		}()
		tlsEnabled = true
		caPath = "capath"
		certFile = "certfile"
		keyFile = "keyfile"

		err := os.MkdirAll("capath", 0755)
		require.NoError(t, err)

		startTraceServer(&TestServer{
			t:                t,
			expectedAddress:  ":7891",
			expectedCertFile: certFile,
			expectedKeyFile:  keyFile,
		}, 7891)
	})

	t.Run("HTTP server when ACE_ADMIN_SERVER_SECURITY is not true", func(t *testing.T) {
		tlsEnabled = false
		startTraceServer(&TestServer{
			t:               t,
			expectedAddress: ":7891",
		}, 7891)
	})
}

func TestUserTraceRouterHandler(t *testing.T) {
	log = testLogger

	handler := http.HandlerFunc(userTraceRouterHandler)
	url := "/collect-user-trace"

	t.Run("Sends an error if not a POST request", func(t *testing.T) {
		request, _ := http.NewRequest("GET", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)
		assert.Equal(t, http.StatusMethodNotAllowed, response.Code)
	})

	t.Run("Sends an error if the trace can't be collected", func(t *testing.T) {
		traceDir = "test/trace"
		request, _ := http.NewRequest("POST", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("Streams a zip file for user trace", func(t *testing.T) {
		defer restoreUserTrace()
		setUpUserTrace(t)

		request, _ := http.NewRequest("POST", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)

		body, _ := io.ReadAll(response.Body)
		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		require.NoError(t, err)
		
		files := checkZip(t, zipReader)
		assert.Len(t, files, 1)
		assert.Contains(t, files, "test.userTrace.txt")
	})
}

func TestServiceTraceRouteHandler(t *testing.T) {
	log = testLogger

	handler := http.HandlerFunc(serviceTraceRouterHandler)
	url := "/collect-service-trace"

	t.Run("Sends an error if not a POST request", func(t *testing.T) {
		request, _ := http.NewRequest("GET", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)
		assert.Equal(t, http.StatusMethodNotAllowed, response.Code)
	})

	t.Run("Sends an error if the trace can't be collected", func(t *testing.T) {
		traceDir = "test/trace"

		request, _ := http.NewRequest("POST", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("Streams a zip file for service trace", func(t *testing.T) {
		defer restoreServiceTrace()
		setUpServiceTrace(t)

		request, _ := http.NewRequest("POST", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)

		body, _ := io.ReadAll(response.Body)
		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		require.NoError(t, err)

		files := checkZip(t, zipReader)
		assert.Len(t, files, 6)
		assert.Contains(t, files, "test.trace.txt")
		assert.Contains(t, files, "test.exceptionLog.txt")
		assert.Contains(t, files, "test.designerflows.txt")
		assert.Contains(t, files, "test.designereventflows.txt")
		assert.Contains(t, files, "env.txt")
		assert.Contains(t, files, "ps -ewww.txt")
	})
}

func TestZipUserTrace(t *testing.T) {
	log = testLogger

	defer restoreUserTrace()
	setUpUserTrace(t)

	t.Run("Builds a zip with user trace", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)

		err := zipUserTrace(zipWriter)
		require.NoError(t, err)

		zipWriter.Close()

		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)
		filesInZip := checkZip(t, zipReader)
		assert.Len(t, filesInZip, 1)
		assert.Contains(t, filesInZip, "test.userTrace.txt")
	})

	t.Run("Returns an error when it can't collect user trace", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := FailOnFileNameZipWriter{
			failOnFileName: "test.userTrace.txt",
			zipWriter:      zip.NewWriter(&buffer),
		}
		err := zipUserTrace(zipWriter)
		assert.EqualError(t, err, "Unable to write test.userTrace.txt")
	})
}

func TestZipServiceTrace(t *testing.T) {
	log = testLogger

	defer restoreServiceTrace()
	setUpServiceTrace(t)

	t.Run("Builds a zip with service trace, exception logs, designer operational logs, env, and ps -ewww output", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)

		err := zipServiceTrace(zipWriter)
		require.NoError(t, err)

		zipWriter.Close()

		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)
		files := checkZip(t, zipReader)
		assert.Len(t, files, 6)
		assert.Contains(t, files, "test.trace.txt")
		assert.Contains(t, files, "test.exceptionLog.txt")
		assert.Contains(t, files, "test.designerflows.txt")
		assert.Contains(t, files, "test.designereventflows.txt")
		assert.Contains(t, files, "env.txt")
		assert.Contains(t, files, "ps -ewww.txt")
	})

	t.Run("Failure test cases", func(t *testing.T) {
		failureTestCases := []string{
			"test.trace.txt",
			"test.designerflows.txt",
			"env.txt",
			"ps -ewww.txt",
		}

		for _, fileName := range failureTestCases {
			t.Run(fileName, func(t *testing.T) {
				var buffer bytes.Buffer
				zipWriter := FailOnFileNameZipWriter{
					failOnFileName: fileName,
					zipWriter:      zip.NewWriter(&buffer),
				}
				err := zipServiceTrace(zipWriter)
				assert.EqualError(t, err, "Unable to write "+fileName)
			})
		}
	})
}

func TestRunOSCommand(t *testing.T) {
	log = testLogger

	t.Run("Returns an error if it can't run the command", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)
		err := runOSCommand(zipWriter, "file.txt", "asdasdd")
		assert.Error(t, err)
		zipWriter.Close()
		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)
		assert.Len(t, checkZip(t, zipReader), 0)
	})

	t.Run("Returns an error if there is an error when creating a file in the zip", func(t *testing.T) {
		zipWriter := CreateFailureZipWriter{}
		err := runOSCommand(zipWriter, "file.txt", "echo", "hello world")
		assert.Error(t, err)
	})

	t.Run("Returns an error if the command output can't be written to the zip", func(t *testing.T) {
		zipWriter := WriteFailureZipWriter{}
		err := runOSCommand(zipWriter, "file.txt", "echo", "hello world")
		assert.Error(t, err)
	})

	t.Run("Adds the command output to the zip if successful", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)

		err := runOSCommand(zipWriter, "file.txt", "echo", "hello world")
		assert.NoError(t, err)

		zipWriter.Close()

		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)

		files := checkZip(t, zipReader)
		assert.Len(t, files, 1)
		assert.Contains(t, files, "file.txt")
	})
}

func TestZipDir(t *testing.T) {
	log = testLogger

	// Create directories and files which will be archived
	err := os.MkdirAll("subdir/parent/child", 0755)
	require.NoError(t, err)
	defer os.RemoveAll("subdir")

	files := []string{"subdir/parent/file1.txt", "subdir/parent/child/file2.txt", "subdir/parent/file3.txt"}

	for _, fileName := range files {
		file, err := os.Create(fileName)
		require.NoError(t, err)
		_, err = file.WriteString("This is a test")
		require.NoError(t, err)
	}

	t.Run("Calls zipFile for each file and only adds files which pass the test function ", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)
		err := zipDir("subdir", zipWriter, func(fileName string) bool {
			return !strings.Contains(fileName, "1")
		})
		zipWriter.Close()
		require.NoError(t, err)

		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)

		files := checkZip(t, zipReader)
		assert.Len(t, files, 2)
		assert.Contains(t, files, "file2.txt")
		assert.Contains(t, files, "file3.txt")
	})

	t.Run("Returns an error if the directory does not exist", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)
		err := zipDir("does-not-exist", zipWriter, func(string) bool { return true })
		zipWriter.Close()
		assert.Error(t, err)
		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)
		assert.Len(t, checkZip(t, zipReader), 0)
	})

	t.Run("Returns an error if passed a file that is not a directory", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)
		err := zipDir("subdir/parent/file1.txt", zipWriter, func(string) bool { return true })
		zipWriter.Close()
		assert.EqualError(t, err, "subdir/parent/file1.txt is not a directory")
		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)
		assert.Len(t, checkZip(t, zipReader), 0)
	})

	t.Run("Creates an empty zip if there are no files which pass the test function", func(t *testing.T) {
		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)
		err := zipDir("subdir", zipWriter, func(fileName string) bool {
			return !strings.Contains(fileName, "file")
		})
		zipWriter.Close()
		assert.NoError(t, err)
		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)
		assert.Len(t, checkZip(t, zipReader), 0)
	})
}

func TestZipFile(t *testing.T) {
	log = testLogger
	fileNameToZip := "fileToZip.txt"

	testSetup := func() {
		// Create file which will be archived
		fileToZip, err := os.Create(fileNameToZip)
		require.NoError(t, err)
		_, err = fileToZip.WriteString("This is a test")
		require.NoError(t, err)
	}

	t.Run("Returns an error when the file cannot be opened", func(t *testing.T) {
		err := zipFile("badPath", nil)
		assert.Error(t, err)
	})

	t.Run("Returns an error when it fails to create the file in the zip", func(t *testing.T) {
		defer os.Remove(fileNameToZip)
		testSetup()

		zipWriter := CreateFailureZipWriter{}

		err := zipFile(fileNameToZip, zipWriter)
		assert.EqualError(t, err, "Failed to create")
	})

	t.Run("Returns an error when it fails to add the file to the zip", func(t *testing.T) {
		defer os.Remove(fileNameToZip)
		testSetup()

		zipWriter := WriteFailureZipWriter{}

		err := zipFile(fileNameToZip, zipWriter)
		assert.EqualError(t, err, "Failed to write")
	})

	t.Run("Returns with no error when the header and file are successfully written to the zip", func(t *testing.T) {
		defer os.Remove(fileNameToZip)
		testSetup()

		var buffer bytes.Buffer
		zipWriter := zip.NewWriter(&buffer)
		err := zipFile("fileToZip.txt", zipWriter)
		assert.NoError(t, err)
		zipWriter.Close()

		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		require.NoError(t, err)
		files := checkZip(t, zipReader)
		assert.Equal(t, 1, len(files))
		assert.Contains(t, files, "fileToZip.txt")
	})
}

func TestReadBasicAuthCreds(t *testing.T) {
	credsDir := "credentials"

	setupDir := func() {
		err := os.MkdirAll(credsDir, 0755)
		require.NoError(t, err)
	}

	cleanUpDir := func() {
		err := os.RemoveAll(credsDir)
		require.NoError(t, err)
	}

	t.Run("Returns an error if the directory does not exist", func(t *testing.T) {
		err := readBasicAuthCreds("does/not/exist", &FileReader{})
		require.Error(t, err)
	})

	t.Run("Returns an error if the input parameter is not a directory", func(t *testing.T) {
		setupDir()
		defer cleanUpDir()

		_, err := os.Create(credsDir + "/emptyFile.txt")
		require.NoError(t, err)

		err = readBasicAuthCreds(credsDir+"/emptyFile.txt", &FileReader{})
		require.EqualError(t, err, "credentials/emptyFile.txt is not a directory")
	})

	t.Run("Returns an error if there is an error reading a file", func(t *testing.T) {
		setupDir()
		defer cleanUpDir()

		_, err := os.Create(credsDir + "/admin-users.txt")
		require.NoError(t, err)

		err = readBasicAuthCreds(credsDir, &ErrorFileReader{})
		require.EqualError(t, err, "Unable to read file")
	})

	t.Run("Returns an error if it fails to parse the credentials file", func(t *testing.T) {
		setupDir()
		defer cleanUpDir()

		file, err := os.Create(credsDir + "/admin-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("This is a test")
		require.NoError(t, err)

		err = readBasicAuthCreds(credsDir, &FileReader{})
		require.EqualError(t, err, "Unable to parse admin-users.txt")
	})

	t.Run("Returns nil if the credentials have all been read and parsed - single line files", func(t *testing.T) {
		log = testLogger

		setupDir()
		defer cleanUpDir()

		credentials = []Credential{}

		file, err := os.Create(credsDir + "/admin-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("user1 pass1")
		require.NoError(t, err)

		file, err = os.Create(credsDir + "/operator-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("user2 pass2")
		require.NoError(t, err)

		err = readBasicAuthCreds(credsDir, &FileReader{})
		require.NoError(t, err)

		require.Len(t, credentials, 2)
		assert.Equal(t, sha256.Sum256([]byte("user1")), credentials[0].Username)
		assert.Equal(t, sha256.Sum256([]byte("pass1")), credentials[0].Password)
		assert.Equal(t, sha256.Sum256([]byte("user2")), credentials[1].Username)
		assert.Equal(t, sha256.Sum256([]byte("pass2")), credentials[1].Password)
	})

	t.Run("Returns nil if the credentials have all been read and parsed - multi line files, trailing spaces and comments", func(t *testing.T) {
		log = testLogger

		setupDir()
		defer cleanUpDir()

		credentials = []Credential{}

		file, err := os.Create(credsDir + "/admin-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("user1 pass1\nuser2 pass2\n")
		require.NoError(t, err)

		file, err = os.Create(credsDir + "/operator-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("# this shouldn't cause an error \n# nor should this\nuser3 pass3 \n ")
		require.NoError(t, err)

		err = readBasicAuthCreds(credsDir, &FileReader{})
		require.NoError(t, err)

		require.Len(t, credentials, 3)
		assert.Equal(t, sha256.Sum256([]byte("user1")), credentials[0].Username)
		assert.Equal(t, sha256.Sum256([]byte("pass1")), credentials[0].Password)
		assert.Equal(t, sha256.Sum256([]byte("user2")), credentials[1].Username)
		assert.Equal(t, sha256.Sum256([]byte("pass2")), credentials[1].Password)
		assert.Equal(t, sha256.Sum256([]byte("user3")), credentials[2].Username)
		assert.Equal(t, sha256.Sum256([]byte("pass3")), credentials[2].Password)
	})

	t.Run("does not add viewer, auditor or editor users", func(t *testing.T) {
		log = testLogger

		setupDir()
		defer cleanUpDir()

		credentials = []Credential{}

		file, err := os.Create(credsDir + "/auditor-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("user1 pass1")
		require.NoError(t, err)

		file, err = os.Create(credsDir + "/operator-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("user2 pass2")
		require.NoError(t, err)

		file, err = os.Create(credsDir + "/editor-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("user3 pass3")
		require.NoError(t, err)

		file, err = os.Create(credsDir + "/viewer-users.txt")
		require.NoError(t, err)
		_, err = file.WriteString("user4 pass4")
		require.NoError(t, err)

		err = readBasicAuthCreds(credsDir, &FileReader{})
		require.NoError(t, err)

		require.Len(t, credentials, 1)
		assert.Equal(t, sha256.Sum256([]byte("user2")), credentials[0].Username)
		assert.Equal(t, sha256.Sum256([]byte("pass2")), credentials[0].Password)
	})
}

func TestBasicAuthMiddlware(t *testing.T) {
	setCredentials := func() {
		credentials = []Credential{{
			Username: sha256.Sum256([]byte("user1")),
			Password: sha256.Sum256([]byte("pass1")),
		}, {
			Username: sha256.Sum256([]byte("user2")),
			Password: sha256.Sum256([]byte("pass2")),
		},
		}
	}

	t.Run("No credentials defined", func(t *testing.T) {
		credentials = []Credential{}
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://testing", nil)

		handlerToTest := basicAuthMiddlware(nil)
		handlerToTest.ServeHTTP(response, req)

		assert.Equal(t, 401, response.Result().StatusCode)
	})

	t.Run("No basic auth credentials in request", func(t *testing.T) {
		setCredentials()
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://testing", nil)

		handlerToTest := basicAuthMiddlware(nil)
		handlerToTest.ServeHTTP(response, req)

		assert.Equal(t, 401, response.Result().StatusCode)
	})

	t.Run("Invalid basic auth credentials", func(t *testing.T) {
		setCredentials()
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://testing", nil)
		req.SetBasicAuth("invaliduser", "invalidpass")

		handlerToTest := basicAuthMiddlware(nil)
		handlerToTest.ServeHTTP(response, req)

		assert.Equal(t, 401, response.Result().StatusCode)
	})

	t.Run("Matching username", func(t *testing.T) {
		setCredentials()
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://testing", nil)
		req.SetBasicAuth("user1", "invalidpass")

		handlerToTest := basicAuthMiddlware(nil)
		handlerToTest.ServeHTTP(response, req)

		assert.Equal(t, 401, response.Result().StatusCode)
	})

	t.Run("Matching password", func(t *testing.T) {
		setCredentials()
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://testing", nil)
		req.SetBasicAuth("invaliduser", "pass1")

		handlerToTest := basicAuthMiddlware(nil)
		handlerToTest.ServeHTTP(response, req)

		assert.Equal(t, 401, response.Result().StatusCode)
	})

	t.Run("Matches first credentials", func(t *testing.T) {
		setCredentials()
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://testing", nil)
		req.SetBasicAuth("user1", "pass1")

		var called bool

		handlerToTest := basicAuthMiddlware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))
		handlerToTest.ServeHTTP(response, req)

		assert.True(t, called)
	})

	t.Run("Matches second credentials", func(t *testing.T) {
		setCredentials()
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://testing", nil)
		req.SetBasicAuth("user2", "pass2")

		var called bool

		handlerToTest := basicAuthMiddlware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))
		handlerToTest.ServeHTTP(response, req)

		assert.True(t, called)
	})
}
func TestGetCACertPool(t *testing.T) {
	log = testLogger

	defer func() {
		os.RemoveAll("cert-dir")
	}()

	err := os.MkdirAll("cert-dir", 0755)
	require.NoError(t, err)

	_, err = os.Create("cert-dir/empty-crt.pem") // empty file for parsing errors
	require.NoError(t, err)

	file, err := os.Create("cert-dir/valid-crt.pem")
	require.NoError(t, err)

	cert := createValidCA()
	_, err = file.Write(cert)
	require.NoError(t, err)

	t.Run("Returns an error if caPath is not a file", func(t *testing.T) {
		_, err := getCACertPool("does-not-exist", &FileReader{})
		require.Error(t, err)
	})

	t.Run("Returns an error if the ca file cannot be read", func(t *testing.T) {
		_, err := getCACertPool("cert-dir/valid-crt.pem", &ErrorFileReader{})
		require.Error(t, err)
	})

	t.Run("Returns an error if the ca file cannot be parsed", func(t *testing.T) {
		_, err := getCACertPool("cert-dir/empty-crt.pem", &FileReader{})
		require.Error(t, err)
	})

	t.Run("Returns a caPool with one cert if the ca file is added successfully", func(t *testing.T) {
		caCertPool, err := getCACertPool("cert-dir/valid-crt.pem", &FileReader{})
		require.NoError(t, err)
		assert.Len(t, caCertPool.Subjects(), 1)
	})

	t.Run("Does not return an error when a file reading error occurs in a directory", func(t *testing.T) {
		caCertPool, err := getCACertPool("cert-dir", &ErrorFileReader{})
		require.NoError(t, err)
		assert.Len(t, caCertPool.Subjects(), 0) // both files won't be read
	})

	t.Run("Does not return an error when files can't be parsed", func(t *testing.T) {
		caCertPool, err := getCACertPool("cert-dir", &FileReader{})
		require.NoError(t, err)
		assert.Len(t, caCertPool.Subjects(), 1) // the empty file will fail to parse
	})
}

func checkZip(t *testing.T, zipReader *zip.Reader) []string {
	var files []string
	for _, f := range zipReader.File {
		files = append(files, f.FileHeader.Name)
	}
	return files
}

func setUpUserTrace(t *testing.T) {
	traceDir = "test/trace"

	err := os.MkdirAll(traceDir, 0755)
	require.NoError(t, err)

	files := []string{
		traceDir + "/test.userTrace.txt",
		traceDir + "/no-match.txt",
	}

	for _, fileName := range files {
		file, err := os.Create(fileName)
		require.NoError(t, err)
		_, err = file.WriteString("This is a test")
		require.NoError(t, err)
	}
}

func restoreUserTrace() {
	os.RemoveAll("test")
}

func setUpServiceTrace(t *testing.T) {
	traceDir = "test/trace"
	operationalLogDir = "test/log"

	directories := []string{
		traceDir,
		operationalLogDir,
	}

	for _, dir := range directories {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	files := []string{
		traceDir + "/test.trace.txt",
		traceDir + "/test.exceptionLog.txt",
		traceDir + "/no-match.txt",
		operationalLogDir + "/test.designerflows.txt",
		operationalLogDir + "/test.designereventflows.txt",
		operationalLogDir + "/no-match.txt",
	}

	for _, fileName := range files {
		file, err := os.Create(fileName)
		require.NoError(t, err)
		_, err = file.WriteString("This is a test")
		require.NoError(t, err)
	}
}

func restoreServiceTrace() {
	os.RemoveAll("test")
}

func createValidCA() []byte {
	privKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	ca := &x509.Certificate{
		SerialNumber: &big.Int{},
		IsCA:         true,
	}

	caBytes, _ := x509.CreateCertificate(rand.Reader, ca, ca, &privKey.PublicKey, privKey)

	caPEM := new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	return caPEM.Bytes()
}

type CreateFailureZipWriter struct{}

func (zw CreateFailureZipWriter) Create(filename string) (io.Writer, error) {
	return nil, errors.New("Failed to create")
}
func (zw CreateFailureZipWriter) Close() error {
	return nil
}

type WriteFailureZipWriter struct{}

func (zw WriteFailureZipWriter) Create(filename string) (io.Writer, error) {
	zipEntry := WriteFailureZipEntry{}
	return zipEntry, nil
}
func (zw WriteFailureZipWriter) Close() error {
	return nil
}

type WriteFailureZipEntry struct{}

func (ze WriteFailureZipEntry) Write(p []byte) (n int, err error) {
	return 0, errors.New("Failed to write")
}

type FailOnFileNameZipWriter struct {
	failOnFileName string
	zipWriter      ZipWriterInterface
}

func (zw FailOnFileNameZipWriter) Create(filename string) (io.Writer, error) {
	if filename == zw.failOnFileName {
		return nil, errors.New("Unable to write " + zw.failOnFileName)
	}
	return zw.zipWriter.Create(filename)
}
func (zw FailOnFileNameZipWriter) Close() error {
	return zw.zipWriter.Close()
}

type ErrorFileReader struct{}

func (fr *ErrorFileReader) ReadFile(string) ([]byte, error) {
	return nil, errors.New("Unable to read file")
}

type TestServer struct {
	t                *testing.T
	expectedAddress  string
	expectedCertFile string
	expectedKeyFile  string
}

func (s *TestServer) Start(address string, mux *http.ServeMux) {
	assert.Equal(s.t, s.expectedAddress, address)
}

func (s *TestServer) StartTLS(address string, mux *http.ServeMux, caCertPool *x509.CertPool, certFile string, keyFile string) {
	assert.Equal(s.t, s.expectedAddress, address)
	assert.Equal(s.t, s.expectedCertFile, certFile)
	assert.Equal(s.t, s.expectedKeyFile, keyFile)
}
