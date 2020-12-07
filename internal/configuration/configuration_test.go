package configuration

import (
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ot4i/ace-docker/internal/logger"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

const testBaseDir = "/tmp/tests"

var testSecretValue = []byte("test secret")

const secretName = "setdbparms.txt-wb2dc"

var testLogger, err = logger.NewLogger(os.Stdout, true, true, "test")

var osMkdirAllRestore = osMkdirAll
var ioutilWriteFileRestore = ioutilWriteFile
var ioutilReadFileRestore = ioutilReadFile
var getSecretRestore = getSecret
var setupClientsRestore = setupClients
var getAllConfigurationsRestore = getAllConfigurations
var osOpenFileRestore = osOpenFile
var ioCopyRestore = ioCopy
var RunCommandRestore = internalRunSetdbparmsCommand
var RunKeytoolCommandRestore = internalRunKeytoolCommand

const policyProjectContent = "UEsDBAoAAAAAAFelclAAAAAAAAAAAAAAAAAJABwAcHJvamVjdDEvVVQJAAPFh3JeyIdyXnV4CwABBPUBAAAEFAAAAFBLAwQUAAAACABgpHJQn5On0w0AAAARAAAAEQAcAHByb2plY3QxL3Rlc3QueG1sVVQJAAPzhXJeHoZyXnV4CwABBPUBAAAEFAAAALMpSS0usQMRNvpgJgBQSwECHgMKAAAAAABXpXJQAAAAAAAAAAAAAAAACQAYAAAAAAAAABAA7UEAAAAAcHJvamVjdDEvVVQFAAPFh3JedXgLAAEE9QEAAAQUAAAAUEsBAh4DFAAAAAgAYKRyUJ+Tp9MNAAAAEQAAABEAGAAAAAAAAQAAAKSBQwAAAHByb2plY3QxL3Rlc3QueG1sVVQFAAPzhXJedXgLAAEE9QEAAAQUAAAAUEsFBgAAAAACAAIApgAAAJsAAAAAAA=="
const genericContent = "UEsDBAoAAAAAAFelclAAAAAAAAAAAAAAAAAJABwAcHJvamVjdDEvVVQJAAPFh3JeyIdyXnV4CwABBPUBAAAEFAAAAFBLAwQUAAAACABgpHJQn5On0w0AAAARAAAAEQAcAHByb2plY3QxL3Rlc3QueG1sVVQJAAPzhXJeHoZyXnV4CwABBPUBAAAEFAAAALMpSS0usQMRNvpgJgBQSwECHgMKAAAAAABXpXJQAAAAAAAAAAAAAAAACQAYAAAAAAAAABAA7UEAAAAAcHJvamVjdDEvVVQFAAPFh3JedXgLAAEE9QEAAAQUAAAAUEsBAh4DFAAAAAgAYKRyUJ+Tp9MNAAAAEQAAABEAGAAAAAAAAQAAAKSBQwAAAHByb2plY3QxL3Rlc3QueG1sVVQFAAPzhXJedXgLAAEE9QEAAAQUAAAAUEsFBgAAAAACAAIApgAAAJsAAAAAAA=="
const adminsslcontent = "UEsDBAoAAAAAAD1epVBBsFEJCgAAAAoAAAAGABwAY2EuY3J0VVQJAAPWRLFe1kSxXnV4CwABBPUBAAAEFAAAAGZha2UgY2VydApQSwECHgMKAAAAAAA9XqVQQbBRCQoAAAAKAAAABgAYAAAAAAABAAAApIEAAAAAY2EuY3J0VVQFAAPWRLFedXgLAAEE9QEAAAQUAAAAUEsFBgAAAAABAAEATAAAAEoAAAAAAA=="
const loopbackdatasourcecontent = "UEsDBBQAAAAIALhhL1E16fencwAAAKQAAAAQABwAZGF0YXNvdXJjZXMuanNvblVUCQAD7KFgX+yhYF91eAsAAQT1AQAABBQAAACrVsrNz0vPT0lSsqpWykvMTVWyUoAJxaeklinpKCXn5+WlJpfkFyFJAYUz8otLQCKGRuZ6BkBoqKSjoFSQXwQSNDI3MDTXUUpJLElMSiwGmwk0y8UJpKS0OLUIZhFQMBTIBetMLC4uzy9KgQoHwLi1tVwAUEsDBAoAAAAAAH1hL1EAAAAAAAAAAAAAAAAGABwAbW9uZ28vVVQJAAN9oWBffaFgX3V4CwABBPUBAAAEFAAAAFBLAQIeAxQAAAAIALhhL1E16fencwAAAKQAAAAQABgAAAAAAAEAAACkgQAAAABkYXRhc291cmNlcy5qc29uVVQFAAPsoWBfdXgLAAEE9QEAAAQUAAAAUEsBAh4DCgAAAAAAfWEvUQAAAAAAAAAAAAAAAAYAGAAAAAAAAAAQAO1BvQAAAG1vbmdvL1VUBQADfaFgX3V4CwABBPUBAAAEFAAAAFBLBQYAAAAAAgACAKIAAAD9AAAAAAA="

func restore() {
	osMkdirAll = osMkdirAllRestore
	ioutilWriteFile = ioutilWriteFileRestore
	ioutilReadFile = ioutilReadFileRestore
	getSecret = getSecretRestore
	getAllConfigurations = getAllConfigurationsRestore
	setupClients = setupClientsRestore
	osOpenFile = osOpenFileRestore
	ioCopy = ioCopyRestore
	internalRunSetdbparmsCommand = RunCommandRestore
	internalRunKeytoolCommand = RunKeytoolCommandRestore
}

func reset() {
	// default error mocks
	osMkdirAll = func(path string, perm os.FileMode) error {
		panic("Should be mocked")
	}
	ioutilWriteFile = func(fn string, data []byte, perm os.FileMode) error {
		panic("Should be mocked")
	}
	getSecret = func(basedir string, secretName string) ([]byte, error) {
		panic("Should be mocked")
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {

		panic("Should be mocked")
	}

	osOpenFile = func(name string, flat int, perm os.FileMode) (*os.File, error) {
		panic("Should be mocked")
	}

	ioCopy = func(dst io.Writer, src io.Reader) (written int64, err error) {
		panic("Should be mocked")
	}

	internalRunSetdbparmsCommand = func(log *logger.Logger, command string, params []string) error {
		panic("Should be mocked")
	}

	internalRunKeytoolCommand = func(log *logger.Logger, params []string) error {
		panic("Should be mocked")
	}

	// common mock
	setupClients = func() (dynamic.Interface, error) {
		return nil, nil

	}

	ioutilReadFile = func(filename string) ([]byte, error) {
		return []byte("ace"), nil
	}
}

func TestUnzip(t *testing.T) {
	zipContents, _ := base64.StdEncoding.DecodeString(policyProjectContent)
	osMkdirAll = func(path string, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+"project1", path)
		return nil
	}
	osOpenFile = func(name string, flat int, perm os.FileMode) (*os.File, error) {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+"project1"+string(os.PathSeparator)+"test.xml", name)
		return &os.File{}, nil
	}

	ioCopy = func(dst io.Writer, src io.Reader) (written int64, err error) {
		b, err := ioutil.ReadAll(src)
		assert.Equal(t, "<test>test</test>", string(b))
		return 12, nil
	}
	// test working  unzip file
	assert.Nil(t, unzip(testLogger, testBaseDir, zipContents))
	// test failing ioCopy
	ioCopy = func(dst io.Writer, src io.Reader) (written int64, err error) {
		return 1, errors.New("failed ioCopy")
	}
	{
		err := unzip(testLogger, testBaseDir, zipContents)
		if assert.Error(t, err, "Failed to copy file") {
			assert.Equal(t, errors.New("failed ioCopy"), err)
		}
	}
	// test open file fail
	osOpenFile = func(name string, flat int, perm os.FileMode) (*os.File, error) {
		return &os.File{}, errors.New("failed file open")
	}
	{
		err := unzip(testLogger, testBaseDir, zipContents)
		if assert.Error(t, err, "Failed to copy file") {
			assert.Equal(t, errors.New("failed file open"), err)
		}
	}
	// fail mkdir
	osMkdirAll = func(path string, perm os.FileMode) error {
		return errors.New("failed mkdir")
	}
	{
		err := unzip(testLogger, testBaseDir, zipContents)
		if assert.Error(t, err, "Failed to copy file") {
			assert.Equal(t, errors.New("failed mkdir"), err)
		}
	}
}
func TestSetupConfigurationsFiles(t *testing.T) {
	// Test env ACE_CONFIGURATIONS not being set
	os.Unsetenv("ACE_CONFIGURATIONS")
	reset()
	assert.Nil(t, SetupConfigurationsFiles(testLogger, testBaseDir))
	// Test env ACE_CONFIGURATIONS being set
	os.Setenv("ACE_CONFIGURATIONS", "server.conf.yaml,odbc.ini")
	reset()
	configContents := "ZmFrZUZpbGUK"
	// Test env ACE_CONFIGURATIONS set to an array of two configurations
	reset()
	osMkdirAll = func(path string, perm os.FileMode) error {
		if path != testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"overrides" &&
			path != testBaseDir+string(os.PathSeparator)+workdirName {
			t.Errorf("Incorrect path for mkdir: %s", path)
		}
		return nil
	}
	ioutilWriteFile = func(fn string, data []byte, perm os.FileMode) error {
		if fn != testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"overrides"+string(os.PathSeparator)+"server.conf.yaml" &&
			fn != testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"odbc.ini" {
			t.Errorf("Incorrect path for writing to a file: %s", fn)
		}
		return nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {

		assert.Equal(t, name, secretName)
		return testSecretValue, nil
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, "server.conf.yaml", cn[0])
		assert.Equal(t, "odbc.ini", cn[1])
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "server.conf.yaml",
					},
					"spec": map[string]interface{}{
						"type":     "serverconf",
						"contents": configContents,
					},
				},
			}, {
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "odbc.ini",
					},
					"spec": map[string]interface{}{
						"type":     "odbc",
						"contents": configContents,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFiles(testLogger, testBaseDir))
}

func TestRunCommand(t *testing.T) {
	// check get an error when command is rubbish
	err := runCommand(testLogger, "fakeCommand", []string{"fake"})
	if err == nil {
		assert.Equal(t, errors.New("Command should have failed"), err)
	}
}

func TestSetupConfigurationsFilesInternal(t *testing.T) {

	configContents := "ZmFrZUZpbGUK"
	// Test missing type fails
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "bad.type")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "bad.type",
					},
					"spec": map[string]interface{}{},
				},
			},
		}, nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"bad.type"}, testBaseDir)
		if assert.Error(t, err, "A configuration must has a type") {
			assert.Equal(t, errors.New("A configuration must has a type"), err)
		}
	}
	// Test missing contents fails
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "bad.type")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "bad.type",
					},
					"spec": map[string]interface{}{
						"type": "serverconf",
					},
				},
			},
		}, nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"bad.type"}, testBaseDir)
		if assert.Error(t, err, "A configuration with type: serverconf must has a contents field") {
			assert.Equal(t, errors.New("A configuration with type: serverconf must has a contents field"), err)
		}
	}
	// Test missing secret fails
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "bad.type")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "bad.type",
					},
					"spec": map[string]interface{}{
						"type": "setdbparms",
					},
				},
			},
		}, nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"bad.type"}, testBaseDir)
		if assert.Error(t, err, "A configuration with type: setdbparms must have a secretName field") {
			assert.Equal(t, errors.New("A configuration with type: setdbparms must have a secretName field"), err)
		}
	} // Test secret file is missing
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return nil, errors.New("missing secret file")
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "bad.type")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "bad.type",
					},
					"spec": map[string]interface{}{
						"type":       "setdbparms",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"bad.type"}, testBaseDir)
		if assert.Error(t, err, "missing secret file") {
			assert.Equal(t, errors.New("missing secret file"), err)
		}
	}
	// Test invalid type fails
	reset()
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "bad.type")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "bad.type",
					},
					"spec": map[string]interface{}{
						"type":     "badtype",
						"contents": configContents,
					},
				},
			},
		}, nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"bad.type"}, testBaseDir)
		assert.Equal(t, errors.New("Unknown configuration type"), err)
	}
	// Test base64 decode fails of content
	reset()
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "server.conf.yaml")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "server.conf.yaml",
					},
					"spec": map[string]interface{}{
						"type":     "serverconf",
						"contents": "not base64",
					},
				},
			},
		}, nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"server.conf.yaml"}, testBaseDir)
		if assert.Error(t, err, "Fails to decode") {
			assert.Equal(t, errors.New("Failed to decode contents"), err)
		}
	}

	// Test accounts.yaml
	reset()
	getSecret = func(basdir string, name string) ([]byte, error) {

		assert.Equal(t, name, secretName)
		return testSecretValue, nil
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "accounts-1")

		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "accounts-1",
					},
					"spec": map[string]interface{}{
						"type":       "accounts",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"accounts-1"}, testBaseDir))
	// Test agentx.json
	reset()
	osMkdirAll = func(path string, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config/iibswitch/agentx", path)
		return nil
	}
	ioutilWriteFile = func(fn string, data []byte, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config/iibswitch/agentx"+string(os.PathSeparator)+"agentx.json", fn)
		return nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {

		assert.Equal(t, name, secretName)
		return testSecretValue, nil
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "agentx-1")

		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "agentx-1",
					},
					"spec": map[string]interface{}{
						"type":       "agentx",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"agentx-1"}, testBaseDir))
	// Test agenta.json
	reset()
	osMkdirAll = func(path string, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config/iibswitch/agenta", path)
		return nil
	}
	ioutilWriteFile = func(fn string, data []byte, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config/iibswitch/agenta"+string(os.PathSeparator)+"agenta.json", fn)
		return nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {

		assert.Equal(t, name, secretName)
		return testSecretValue, nil
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "agenta-1")

		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "agenta-1",
					},
					"spec": map[string]interface{}{
						"type":       "agenta",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"agenta-1"}, testBaseDir))
	// Test odbc.ini using contents field
	reset()
	osMkdirAll = func(path string, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName, path)
		return nil
	}
	ioutilWriteFile = func(fn string, data []byte, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"odbc.ini", fn)
		return nil
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "odbc-ini")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "odbc-ini",
					},
					"spec": map[string]interface{}{
						"type":     "odbc",
						"contents": configContents,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"odbc-ini"}, testBaseDir))
	// Test Truststore
	reset()
	osMkdirAll = func(path string, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+"truststores", path)
		return nil
	}
	ioutilWriteFile = func(fn string, data []byte, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+"truststores"+string(os.PathSeparator)+"truststore-1", fn)
		return nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {

		assert.Equal(t, name, secretName)
		return testSecretValue, nil
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "truststore-1")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "truststore-1",
					},
					"spec": map[string]interface{}{
						"type":       "truststore",
						"secretName": secretName},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"truststore-1"}, testBaseDir))
	// Test Keystore
	reset()
	osMkdirAll = func(path string, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+"keystores", path)
		return nil
	}
	ioutilWriteFile = func(fn string, data []byte, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+"keystores"+string(os.PathSeparator)+"keystore-1", fn)
		return nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {

		assert.Equal(t, name, secretName)
		return testSecretValue, nil
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "keystore-1")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "keystore-1",
					},
					"spec": map[string]interface{}{
						"type":       "keystore",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"keystore-1"}, testBaseDir))
	// Test setdbparms.txt
	reset()
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "setdbparms.txt")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "setdbparms.txt",
					},
					"spec": map[string]interface{}{
						"type":       "setdbparms",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return testSecretValue, nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"setdbparms.txt"}, testBaseDir)
		assert.Equal(t, errors.New("Invalid mqsisetdbparms entry - too few parameters"), err)
	}
	// Test setdbparms with too many parameters
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return []byte("name user pass extra"), nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"setdbparms.txt"}, testBaseDir)
		assert.Equal(t, errors.New("Invalid mqsisetdbparms entry - too many parameters"), err)
	}
	// Test setdbparms with just name, user and password but command fails
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return []byte("name user pass"), nil
	}
	internalRunSetdbparmsCommand = func(log *logger.Logger, command string, params []string) error {
		assert.Equal(t, command, "mqsisetdbparms")
		testParams := []string{"'-n'", "'name'", "'-u'", "'user'", "'-p'", "'pass'", "'-w'", "'/tmp/tests/ace-server'"}
		assert.Equal(t, params, testParams)
		return errors.New("command fails")
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"setdbparms.txt"}, testBaseDir)
		assert.Equal(t, errors.New("command fails"), err)
	}
	// Test setdbparms with full command but command fails
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return []byte("mqsisetdbparms    -n name    -u    user     -p        pass    -w    /tmp/tests/ace-server"), nil
	}
	internalRunSetdbparmsCommand = func(log *logger.Logger, command string, params []string) error {
		assert.Equal(t, command, "mqsisetdbparms")
		testParams := []string{"'-n'", "'name'", "'-u'", "'user'", "'-p'", "'pass'", "'-w'", "'/tmp/tests/ace-server'"}
		assert.Equal(t, params, testParams)
		return errors.New("command fails")
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"setdbparms.txt"}, testBaseDir)
		assert.Equal(t, errors.New("command fails"), err)
	}
	// Test setdbparms with just name, user and password
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return []byte("name user pass"), nil
	}
	internalRunSetdbparmsCommand = func(log *logger.Logger, command string, params []string) error {
		assert.Equal(t, command, "mqsisetdbparms")
		testParams := []string{"'-n'", "'name'", "'-u'", "'user'", "'-p'", "'pass'", "'-w'", "'/tmp/tests/ace-server'"}
		assert.Equal(t, params, testParams)
		return nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"setdbparms.txt"}, testBaseDir))

	// Test setdbparms with several lines, spaces and single quotes
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return []byte("\n   name1    user1     pass1   \n  name2    user2     pass2'  "), nil

	}
	internalRunSetdbparmsCommand = func(log *logger.Logger, command string, params []string) error {
		assert.Equal(t, command, "mqsisetdbparms")
		var testParams []string
		if params[1] == "'name1'" {
			testParams = []string{"'-n'", "'name1'", "'-u'", "'user1'", "'-p'", "'pass1'", "'-w'", "'/tmp/tests/ace-server'"}
		} else {
			testParams = []string{"'-n'", "'name2'", "'-u'", "'user2'", "'-p'", "'pass2'\\'''", "'-w'", "'/tmp/tests/ace-server'"}

		}
		assert.Equal(t, params, testParams)
		return nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"setdbparms.txt"}, testBaseDir))
	// Test setdbparms with full syntax
	getSecret = func(basdir string, name string) ([]byte, error) {

		assert.Equal(t, name, secretName)
		return []byte("mqsisetdbparms -n name -u user -p pass"), nil
	}
	internalRunSetdbparmsCommand = func(log *logger.Logger, command string, params []string) error {
		assert.Equal(t, command, "mqsisetdbparms")
		testParams := []string{"'-n'", "'name'", "'-u'", "'user'", "'-p'", "'pass'", "'-w'", "'/tmp/tests/ace-server'"}
		assert.Equal(t, params, testParams)
		return nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"setdbparms.txt"}, testBaseDir))

	// Test setdbparms with spaces and -w included
	getSecret = func(basdir string, name string) ([]byte, error) {

		assert.Equal(t, name, secretName)
		return []byte("mqsisetdbparms    -n name    -u    user     -p        pass    -w    /tmp/tests/ace-server"), nil
	}
	internalRunSetdbparmsCommand = func(log *logger.Logger, command string, params []string) error {
		assert.Equal(t, command, "mqsisetdbparms")
		testParams := []string{"'-n'", "'name'", "'-u'", "'user'", "'-p'", "'pass'", "'-w'", "'/tmp/tests/ace-server'"}
		assert.Equal(t, params, testParams)
		return nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"setdbparms.txt"}, testBaseDir))

	// policy project with an invalid zip file
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "policy-project")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "policy-project",
					},
					"spec": map[string]interface{}{
						"type":     "policyproject",
						"contents": configContents,
					},
				},
			},
		}, nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"policy-project"}, testBaseDir)
		assert.Equal(t, errors.New("zip: not a valid zip file"), err)
	}
	// Test adminssl
	reset()
	osMkdirAll = func(path string, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+adminsslName, path)
		return nil
	}
	osOpenFile = func(name string, flat int, perm os.FileMode) (*os.File, error) {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+adminsslName+string(os.PathSeparator)+"ca.crt", name)
		return &os.File{}, nil
	}

	ioCopy = func(dst io.Writer, src io.Reader) (written int64, err error) {
		b, err := ioutil.ReadAll(src)
		assert.Equal(t, "fake cert\n", string(b))
		return 12, nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return base64.StdEncoding.DecodeString(adminsslcontent)
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "adminssl1")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "adminssl1",
					},
					"spec": map[string]interface{}{
						"type":       "adminssl",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"adminssl1"}, testBaseDir))
	// Test adminssl with invalid zip file
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return []byte("not a zip"), nil
	}

	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"adminssl1"}, testBaseDir)
		assert.Equal(t, errors.New("zip: not a valid zip file"), err)
	}
	reset()
	// Test generic
	osMkdirAll = func(path string, perm os.FileMode) error {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+genericName+string(os.PathSeparator)+"project1", path)
		return nil
	}
	osOpenFile = func(name string, flat int, perm os.FileMode) (*os.File, error) {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+genericName+string(os.PathSeparator)+"project1"+string(os.PathSeparator)+"test.xml", name)
		return &os.File{}, nil
	}

	ioCopy = func(dst io.Writer, src io.Reader) (written int64, err error) {
		b, err := ioutil.ReadAll(src)
		assert.Equal(t, "<test>test</test>", string(b))
		return 12, nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return base64.StdEncoding.DecodeString(genericContent)
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "generic1")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "generic1",
					},
					"spec": map[string]interface{}{
						"type":       "generic",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"generic1"}, testBaseDir))
	// Test generic with invalid zip file
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return []byte("not a zip"), nil
	}

	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"generic1"}, testBaseDir)
		assert.Equal(t, errors.New("zip: not a valid zip file"), err)
	}
	reset()
	// Test loopbackdatasource
	countMkDir := 0
	osMkdirAll = func(path string, perm os.FileMode) error {
		if countMkDir == 0 {
			assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config"+string(os.PathSeparator)+"connectors"+string(os.PathSeparator)+"loopback", path)
		} else {
			assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config"+string(os.PathSeparator)+"connectors"+string(os.PathSeparator)+"loopback"+string(os.PathSeparator)+"mongo", path)
		}
		countMkDir++
		return nil
	}
	osOpenFile = func(name string, flat int, perm os.FileMode) (*os.File, error) {
		assert.Equal(t, testBaseDir+string(os.PathSeparator)+workdirName+string(os.PathSeparator)+"config"+string(os.PathSeparator)+"connectors"+string(os.PathSeparator)+"loopback"+string(os.PathSeparator)+"datasources.json", name)
		return &os.File{}, nil
	}

	ioCopy = func(dst io.Writer, src io.Reader) (written int64, err error) {
		b, err := ioutil.ReadAll(src)
		assert.Equal(t, "{\"mongodb\":{\"name\": \"mongodb_dev\",\"connector\": \"mongodb\",\"host\": \"127.0.0.1\", \"port\": 27017,\"database\": \"devDB\", \"username\": \"devUser\", \"password\": \"devPassword\"}}\n", string(b))
		return 12, nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return base64.StdEncoding.DecodeString(loopbackdatasourcecontent)
	}
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "loopback1")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "loopback1",
					},
					"spec": map[string]interface{}{
						"type":       "loopbackdatasource",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	assert.Nil(t, SetupConfigurationsFilesInternal(testLogger, []string{"loopback1"}, testBaseDir))
	// Test loopbackdatasource with invalid zip file
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return []byte("not a zip"), nil
	}

	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"loopback1"}, testBaseDir)

		assert.Equal(t, errors.New("zip: not a valid zip file"), err)

	}

	// Test truststore certificates
	getAllConfigurations = func(log *logger.Logger, ns string, cn []string, dc dynamic.Interface) ([]*unstructured.Unstructured, error) {
		assert.Equal(t, cn[0], "truststorecert")
		return []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"kind":       "Configurations",
					"apiVersion": "v1",
					"metadata": map[string]interface{}{
						"name": "truststorecert",
					},
					"spec": map[string]interface{}{
						"type":       "truststorecertificate",
						"secretName": secretName,
					},
				},
			},
		}, nil
	}
	getSecret = func(basdir string, name string) ([]byte, error) {
		assert.Equal(t, name, secretName)
		return testSecretValue, nil
	}

	// Test keytool command to be called with the correct params
	internalRunKeytoolCommandCallCount := 0
	internalRunKeytoolCommand = func(log *logger.Logger, params []string) error {
		testParams := []string{"-import", "-file", "-alias", "truststorecert", "-keystore", "$MQSI_JREPATH/lib/security/cacerts", "-storepass", "changeit", "-noprompt", "-storetype", "JKS"}
		for i := range params {
			if i < 2 {
				assert.Equal(t, params[i], testParams[i])
			} else if i > 2 {
				assert.Equal(t, params[i], testParams[i-1])
			}
		}
		internalRunKeytoolCommandCallCount++
		return nil
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"truststorecert"}, testBaseDir)
		assert.Equal(t, 1, internalRunKeytoolCommandCallCount)
		assert.Equal(t, nil, err)
	}

	// Test keytool command throw an error
	internalRunKeytoolCommand = func(log *logger.Logger, params []string) error {
		return errors.New("command fails")
	}
	{
		err := SetupConfigurationsFilesInternal(testLogger, []string{"truststorecert"}, testBaseDir)
		assert.Equal(t, errors.New("command fails"), err)
	}

	// restore
	restore()
}
