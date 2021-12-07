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

// Package trace contains code to collect trace files
package trace

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ot4i/ace-docker/common/logger"
)

var log logger.LoggerInterface

var traceDir = "/home/aceuser/ace-server/config/common/log"
var credsDir = "/home/aceuser/initial-config/webusers"

var tlsEnabled = os.Getenv("ACE_ADMIN_SERVER_SECURITY") == "true"
var caPath = os.Getenv("ACE_ADMIN_SERVER_CA")
var certFile = os.Getenv("ACE_ADMIN_SERVER_CERT")
var keyFile = os.Getenv("ACE_ADMIN_SERVER_KEY")

var credentials []Credential

type includeFile func(string) bool
type zipFunction func(ZipWriterInterface) error

type ReqBody struct {
	Type string
}

type ZipWriterInterface interface {
	Create(string) (io.Writer, error)
	Close() error
}

type FileReaderInterface interface {
	ReadFile(string) ([]byte, error)
}

type FileReader struct{}

func (fr *FileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

type Credential struct {
	Username [32]byte
	Password [32]byte
}

type ServerInterface interface {
	Start(address string, mux *http.ServeMux)
	StartTLS(address string, mux *http.ServeMux, caCertPool *x509.CertPool, certPath string, keyPath string)
}

type Server struct{}

func (s *Server) Start(address string, mux *http.ServeMux) {
	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Println("Tracing: Trace API server terminated with error " + err.Error())
		} else {
			log.Println("Tracing: Trace API server terminated")
		}
	}()
}

func (s *Server) StartTLS(address string, mux *http.ServeMux, caCertPool *x509.CertPool, certFile string, keyFile string) {
	server := &http.Server{
		Addr:    address,
		Handler: mux,
		TLSConfig: &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}
	go func() {
		err := server.ListenAndServeTLS(certFile, keyFile)
		if err != nil {
			log.Println("Tracing: Trace API server terminated with error " + err.Error())
		} else {
			log.Println("Tracing: Trace API server terminated")
		}
	}()
}

func StartServer(logger logger.LoggerInterface, portNumber int) error {
	log = logger

	err := readBasicAuthCreds(credsDir, &FileReader{})
	if err != nil {
		log.Println("Tracing: No web admin users have been found. The trace APIs will be run without credential verification.")
	}
	return startTraceServer(&Server{}, portNumber)
}

func startTraceServer(server ServerInterface, portNumber int) error {
	address := ":" + strconv.Itoa(portNumber)

	mux := http.NewServeMux()
	serviceTraceHandler := http.HandlerFunc(serviceTraceRouterHandler)
	userTraceHandler := http.HandlerFunc(userTraceRouterHandler)
	mux.Handle("/collect-service-trace", basicAuthMiddlware(serviceTraceHandler))
	mux.Handle("/collect-user-trace", basicAuthMiddlware(userTraceHandler))

	if tlsEnabled {
		caCertPool, err := getCACertPool(caPath, &FileReader{})
		if err != nil {
			return err
		}
		server.StartTLS(address, mux, caCertPool, certFile, keyFile)
	} else {
		server.Start(address, mux)
	}
	return nil
}

func userTraceRouterHandler(res http.ResponseWriter, req *http.Request) {
	traceRouteHandler(res, req, zipUserTrace)
}

func serviceTraceRouterHandler(res http.ResponseWriter, req *http.Request) {
	traceRouteHandler(res, req, zipServiceTrace)
}

func traceRouteHandler(res http.ResponseWriter, req *http.Request, zipFunc zipFunction) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	res.Header().Set("Transfer-Encoding", "chunked")
	res.Header().Set("Content-Disposition", "attachment; filename=\"trace.zip\"")

	zipWriter := zip.NewWriter(res)
	defer zipWriter.Close()

	err := zipFunc(zipWriter)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func zipUserTrace(zipWriter ZipWriterInterface) error {
	log.Println("Tracing: Collecting user trace")
	err := zipDir(traceDir, zipWriter, func(fileName string) bool {
		return strings.Contains(fileName, ".userTrace.")
	})
	if err != nil {
		log.Error("Tracing: Failed to collect user trace. Error: " + err.Error())
		return err
	}
	log.Println("Tracing: Finished collecting user trace")
	return nil
}

func zipServiceTrace(zipWriter ZipWriterInterface) error {
	log.Println("Tracing: Collecting service trace")
	err := zipDir(traceDir, zipWriter, func(fileName string) bool {
		return strings.Contains(fileName, ".trace.") || strings.Contains(fileName, ".exceptionLog.")
	})
	if err != nil {
		log.Error("Tracing: Failed to collect service trace and exception logs. Error: " + err.Error())
		return err
	}

	err = addEnvToZip(zipWriter, "env.txt")
	if err != nil {
		log.Error("Tracing: Failed to get integration server env. Error: " + err.Error())
		return err
	}

	_ = runOSCommand(zipWriter, "ps eww.txt", "ps", "eww")

	log.Println("Tracing: Finished collecting service trace")
	return nil
}

func addEnvToZip(zipWriter ZipWriterInterface, filename string) error {
	log.Println("Tracing: Adding environment variables to zip")

	var envVars []byte
	for _, element := range os.Environ() {
		element += "\n"
		envVars = append(envVars, element...)
	}

	err := addEntryToZip(zipWriter, filename, envVars)
	if err != nil {
		log.Error("Tracing: Unable to add env vars to zip. Error: " + err.Error())
		return err
	}

	return nil
}

func runOSCommand(zipWriter ZipWriterInterface, filename string, command string, arg ...string) error {
	log.Println("Tracing: Collecting output of command: " + command)
	cmd := exec.Command(command, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Error("Tracing: Unable to run command " + command + ": " + err.Error())
		return err
	}

	outBytes := out.Bytes()

	err = addEntryToZip(zipWriter, filename, outBytes)
	if err != nil {
		log.Error("Tracing: Unable to add output of command: " + command + " to zip. Error: " + err.Error())
		return err
	}

	return nil
}

func addEntryToZip(zipWriter ZipWriterInterface, filename string, fileContents []byte) error {
	zipEntry, err := zipWriter.Create(filename)
	if err != nil {
		log.Error("Tracing: Failed to write header for " + filename)
		return err
	}

	if _, err := zipEntry.Write(fileContents); err != nil {
		log.Error("Tracing: Failed to add " + filename + " to archive")
		return err
	}

	return nil
}

func zipDir(traceDir string, zipWriter ZipWriterInterface, testFunc includeFile) error {
	log.Println("Tracing: Creating archive of " + traceDir)
	stat, err := os.Stat(traceDir)
	if err != nil {
		log.Error("Tracing: Directory " + traceDir + " does not exist")
		return err
	}

	if !stat.Mode().IsDir() {
		log.Error("Tracing: " + traceDir + " is not a directory")
		return errors.New(traceDir + " is not a directory")
	}

	return filepath.Walk(traceDir, func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.Mode().IsDir() {
			return nil
		}
		if testFunc(fileInfo.Name()) {
			return zipFile(path, zipWriter)
		}
		return nil
	})
}

func zipFile(path string, zipWriter ZipWriterInterface) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if fileInfo, err := file.Stat(); err == nil {
		log.Println("Tracing: Adding " + fileInfo.Name() + " to archive")

		zipEntry, err := zipWriter.Create(fileInfo.Name())
		if err != nil {
			log.Error("Tracing: Failed to write header for " + fileInfo.Name())
			return err
		}

		if _, err := io.Copy(zipEntry, file); err != nil {
			log.Error("Tracing: Failed to add " + fileInfo.Name() + " to archive")
			return err
		}
	}

	return nil
}

func readBasicAuthCreds(credsDir string, fileReader FileReaderInterface) error {
	stat, err := os.Stat(credsDir)
	if err != nil {
		return err
	}

	if !stat.Mode().IsDir() {
		return errors.New(credsDir + " is not a directory")
	}

	return filepath.Walk(credsDir, func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.Mode().IsDir() {
			return nil
		}

		fileName := fileInfo.Name()

		if fileName == "admin-users.txt" || fileName == "operator-users.txt" {
			file, err := fileReader.ReadFile(path)
			if err != nil {
				return err
			}

			fileString := strings.TrimSpace(string(file))

			lines := strings.Split(fileString, "\n")
			for _, line := range lines {
				if line != "" && !strings.HasPrefix(line, "#") {
					fields := strings.Fields(line)
					if len(fields) != 2 {
						return errors.New("Tracing: Unable to parse " + fileName)
					}
					// using hashes means that the length of the byte array to compare is always the same
					credentials = append(credentials, Credential{
						Username: sha256.Sum256([]byte(fields[0])),
						Password: sha256.Sum256([]byte(fields[1])),
					})
				}
			}

			log.Println("Tracing: Added credentials from " + fileName + " to trace router")
		}
		return nil
	})
}

func basicAuthMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if len(credentials) > 0 {
			username, password, ok := req.BasicAuth()

			if ok {
				usernameHash := sha256.Sum256([]byte(username))
				passwordHash := sha256.Sum256([]byte(password))

				for _, credential := range credentials {
					// subtle.ConstantTimeCompare takes the same amount of time to run, regardless of whether the slices match or not
					usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], credential.Username[:])
					passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], credential.Password[:])
					if usernameMatch+passwordMatch == 2 {
						next.ServeHTTP(res, req)
						return
					}
				}
			}

			http.Error(res, "Unauthorized", http.StatusUnauthorized)
		} else {
			next.ServeHTTP(res, req)
		}
	})
}

func getCACertPool(caPath string, fileReader FileReaderInterface) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()

	stat, err := os.Stat(caPath)

	if err != nil {
		log.Printf("Tracing: %s does not exist", caPath)
		return nil, err
	}

	if stat.IsDir() {
		// path is a directory load all certs
		log.Printf("Tracing: Using CA Certificate folder %s", caPath)
		filepath.Walk(caPath, func(cert string, info os.FileInfo, err error) error {
			if strings.HasSuffix(cert, "crt.pem") {
				log.Printf("Tracing: Adding Certificate %s to CA pool", cert)
				binaryCert, err := fileReader.ReadFile(cert)
				if err != nil {
					log.Printf("Tracing: Error reading CA Certificate %s", err.Error())
					return nil
				}
				ok := caCertPool.AppendCertsFromPEM(binaryCert)
				if !ok {
					log.Printf("Tracing: Failed to parse Certificate %s", cert)
				}
			}
			return nil
		})
	} else {
		log.Printf("Tracing: Using CA Certificate file %s", caPath)
		caCert, err := fileReader.ReadFile(caPath)
		if err != nil {
			log.Errorf("Tracing: Error reading CA Certificate %s", err)
			return nil, err
		}
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			log.Error("Tracing: Failed to parse root CA Certificate")
			return nil, errors.New("failed to parse root CA Certificate")
		}
	}

	return caCertPool, nil
}
