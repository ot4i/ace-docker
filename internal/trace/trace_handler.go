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
var operationalLogDir = "/home/aceuser/ace-server/log"
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
			log.Println("Trace API server terminated with error " + err.Error())
		} else {
			log.Println("Trace API server terminated")
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
			log.Println("Trace API server terminated with error " + err.Error())
		} else {
			log.Println("Trace API server terminated")
		}
	}()
}

func StartServer(logger logger.LoggerInterface, portNumber int) error {
	log = logger

	err := readBasicAuthCreds(credsDir, &FileReader{})
	if err != nil {
		log.Println("Failed to read basic auth credentials. Error: " + err.Error())
		return err
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
	err := zipDir(traceDir, zipWriter, func(fileName string) bool {
		return strings.Contains(fileName, ".userTrace.")
	})
	if err != nil {
		log.Error("Failed to collect user trace. Error: " + err.Error())
		return err
	}

	return nil
}

func zipServiceTrace(zipWriter ZipWriterInterface) error {
	err := zipDir(traceDir, zipWriter, func(fileName string) bool {
		return strings.Contains(fileName, ".trace.") || strings.Contains(fileName, ".exceptionLog.")
	})
	if err != nil {
		log.Error("Failed to collect service trace and exception logs. Error: " + err.Error())
		return err
	}

	err = zipDir(operationalLogDir, zipWriter, func(fileName string) bool {
		return strings.Contains(fileName, ".designerflows.") || strings.Contains(fileName, ".designereventflows.")
	})
	if err != nil {
		log.Error("Failed to collect designer operational logs. Error: " + err.Error())
		return err
	}

	err = runOSCommand(zipWriter, "env.txt", "env")
	if err != nil {
		log.Error("Failed to get integration server env. Error: " + err.Error())
		return err
	}

	err = runOSCommand(zipWriter, "ps -ewww.txt", "ps", "-ewww")
	if err != nil {
		log.Error("Failed to get integration server env. Error: " + err.Error())
		return err
	}

	return nil
}

func runOSCommand(zipWriter ZipWriterInterface, filename string, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Error("Unable to run command " + command + ": " + err.Error())
		return err
	}

	outBytes := out.Bytes()

	zipEntry, err := zipWriter.Create(filename)
	if err != nil {
		log.Error("Failed to write header for " + filename)
		return err
	}

	if _, err := zipEntry.Write(outBytes); err != nil {
		log.Error("Failed to add " + filename + " to archive")
		return err
	}

	return nil
}

func zipDir(traceDir string, zipWriter ZipWriterInterface, testFunc includeFile) error {
	log.Println("Creating archive of " + traceDir)
	stat, err := os.Stat(traceDir)
	if err != nil {
		log.Error("Directory " + traceDir + " does not exist")
		return err
	}

	if !stat.Mode().IsDir() {
		log.Error(traceDir + " is not a directory")
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
		log.Println("Adding " + fileInfo.Name() + " to archive")

		zipEntry, err := zipWriter.Create(fileInfo.Name())
		if err != nil {
			log.Error("Failed to write header for " + fileInfo.Name())
			return err
		}

		if _, err := io.Copy(zipEntry, file); err != nil {
			log.Error("Failed to add " + fileInfo.Name() + " to archive")
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
						return errors.New("Unable to parse " + fileName)
					}
					// using hashes means that the length of the byte array to compare is always the same
					credentials = append(credentials, Credential{
						Username: sha256.Sum256([]byte(fields[0])),
						Password: sha256.Sum256([]byte(fields[1])),
					})
				}
			}

			log.Println("Added credentials from " + fileName + " to trace router")
		}
		return nil
	})
}

func basicAuthMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
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
	})
}

func getCACertPool(caPath string, fileReader FileReaderInterface) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()

	stat, err := os.Stat(caPath)

	if err != nil {
		log.Printf("%s does not exist", caPath)
		return nil, err
	}

	if stat.IsDir() {
		// path is a directory load all certs
		log.Printf("Using CA Certificate folder %s", caPath)
		filepath.Walk(caPath, func(cert string, info os.FileInfo, err error) error {
			if strings.HasSuffix(cert, "crt.pem") {
				log.Printf("Adding Certificate %s to CA pool", cert)
				binaryCert, err := fileReader.ReadFile(cert)
				if err != nil {
					log.Printf("Error reading CA Certificate %s", err.Error())
					return nil
				}
				ok := caCertPool.AppendCertsFromPEM(binaryCert)
				if !ok {
					log.Printf("Failed to parse Certificate %s", cert)
				}
			}
			return nil
		})
	} else {
		log.Printf("Using CA Certificate file %s", caPath)
		caCert, err := fileReader.ReadFile(caPath)
		if err != nil {
			log.Errorf("Error reading CA Certificate %s", err)
			return nil, err
		}
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			log.Error("Failed to parse root CA Certificate")
			return nil, errors.New("failed to parse root CA Certificate")
		}
	}

	return caCertPool, nil
}
