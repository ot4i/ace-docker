package main

import (
	"errors"
	"os"
	"testing"

	"github.com/ot4i/ace-docker/common/logger"
	"github.com/stretchr/testify/assert"
)

func Test_initialIntegrationServerConfig(t *testing.T) {
	oldGetConfigurationFromContentServer := getConfigurationFromContentServer
	getConfigurationFromContentServer = func() error {
		return nil
	}
	t.Run("Golden path - When initial-config/webusers dir exists we call into ConfigureWebAdminUsers to process users", func(t *testing.T) {
		oldCreateSHAServerConfYaml := createSHAServerConfYaml
		createSHAServerConfYaml = func() error {
			return nil
		}
		oldConfigureWebAdminUsers := ConfigureWebAdminUsers
		ConfigureWebAdminUsers = func(log logger.LoggerInterface) error {
			return nil
		}

		homedir = "../../internal/webadmin/testdata"
		initialConfigDir = "../../internal/webadmin/testdata/initial-config"
		err := initialIntegrationServerConfig()
		assert.NoError(t, err)

		createSHAServerConfYaml = oldCreateSHAServerConfYaml
		ConfigureWebAdminUsers = oldConfigureWebAdminUsers

	})

	t.Run("When we fail to properly configure WebAdmin users we fail and return error", func(t *testing.T) {
		oldCreateSHAServerConfYaml := createSHAServerConfYaml
		createSHAServerConfYaml = func() error {
			return nil
		}
		oldConfigureWebAdminUsers := ConfigureWebAdminUsers
		ConfigureWebAdminUsers = func(log logger.LoggerInterface) error {
			return errors.New("Error processing WebAdmin users")
		}
		homedir = "../../internal/webadmin/testdata"
		initialConfigDir = "../../internal/webadmin/testdata/initial-config"
		err := initialIntegrationServerConfig()
		assert.Error(t, err)
		assert.Equal(t, "Error processing WebAdmin users", err.Error())

		createSHAServerConfYaml = oldCreateSHAServerConfYaml
		ConfigureWebAdminUsers = oldConfigureWebAdminUsers
	})

	getConfigurationFromContentServer = oldGetConfigurationFromContentServer
}

func Test_createSHAServerConfYamlLocal(t *testing.T) {
	t.Run("Golden path - Empty file gets populated with the right values", func(t *testing.T) {

		oldReadServerConfFile := readServerConfFile
		readServerConfFile = func() ([]byte, error) {
			return []byte{}, nil
		}
		oldWriteServerConfFileLocal := writeServerConfFile
		writeServerConfFile = func(content []byte) error {
			return nil
		}

		err := createSHAServerConfYaml()
		assert.NoError(t, err)
		readServerConfFile = oldReadServerConfFile
		writeServerConfFile = oldWriteServerConfFileLocal
	})
	t.Run("Golden path 2 - Populated file gets handled and no errors", func(t *testing.T) {

		oldReadServerConfFile := readServerConfFile
		readServerConfFile = func() ([]byte, error) {
			file, err := os.ReadFile("../../internal/webadmin/testdata/initial-config/webusers/server.conf.yaml")
			if err != nil {
				t.Log(err)
				t.Fail()
			}
			return file, nil
		}
		oldWriteServerConfFileLocal := writeServerConfFile
		writeServerConfFile = func(content []byte) error {
			return nil
		}
		oldYamlMarshal := yamlMarshal
		yamlMarshal = func(in interface{}) (out []byte, err error) {
			return nil, nil
		}

		err := createSHAServerConfYaml()
		assert.NoError(t, err)
		readServerConfFile = oldReadServerConfFile
		writeServerConfFile = oldWriteServerConfFileLocal
		yamlMarshal = oldYamlMarshal

	})
	t.Run("Error reading server.conf.yaml file", func(t *testing.T) {
		oldReadServerConfFile := readServerConfFile
		readServerConfFile = func() ([]byte, error) {
			return nil, errors.New("Error reading server.conf.yaml")
		}
		oldWriteServerConfFileLocal := writeServerConfFile
		writeServerConfFile = func(content []byte) error {
			return nil
		}

		err := createSHAServerConfYaml()
		assert.Error(t, err)
		readServerConfFile = oldReadServerConfFile
		writeServerConfFile = oldWriteServerConfFileLocal
	})
	t.Run("yaml.Marshal fails to execute properly", func(t *testing.T) {
		oldYamlUnmarshal := yamlUnmarshal
		yamlUnmarshal = func(in []byte, out interface{}) (err error) {
			return errors.New("Error unmarshalling yaml")
		}
		oldYamlMarshal := yamlMarshal
		yamlMarshal = func(in interface{}) (out []byte, err error) {
			return nil, nil
		}

		err := createSHAServerConfYaml()
		assert.Error(t, err)
		assert.Equal(t, "Error unmarshalling yaml", err.Error())

		yamlUnmarshal = oldYamlUnmarshal
		yamlMarshal = oldYamlMarshal
	})
	t.Run("yaml.Marshal fails to execute properly", func(t *testing.T) {
		oldYamlUnmarshal := yamlUnmarshal
		yamlUnmarshal = func(in []byte, out interface{}) (err error) {
			return nil
		}
		oldYamlMarshal := yamlMarshal
		yamlMarshal = func(in interface{}) (out []byte, err error) {
			return nil, errors.New("Error marshalling yaml")
		}

		err := createSHAServerConfYaml()
		assert.Error(t, err)
		assert.Equal(t, "Error marshalling yaml", err.Error())

		yamlUnmarshal = oldYamlUnmarshal
		yamlMarshal = oldYamlMarshal
	})
	t.Run("yaml.Marshal fails to execute properly", func(t *testing.T) {
		oldYamlUnmarshal := yamlUnmarshal
		yamlUnmarshal = func(in []byte, out interface{}) (err error) {
			return nil
		}
		oldYamlMarshal := yamlMarshal
		yamlMarshal = func(in interface{}) (out []byte, err error) {
			return nil, nil
		}
		oldWriteServerConfFile := writeServerConfFile
		writeServerConfFile = func(content []byte) error {
			return errors.New("Error writing server.conf.yaml")
		}
		err := createSHAServerConfYaml()
		assert.Error(t, err)
		assert.Equal(t, "Error writing server.conf.yaml", err.Error())

		yamlUnmarshal = oldYamlUnmarshal
		yamlMarshal = oldYamlMarshal
		writeServerConfFile = oldWriteServerConfFile
	})
}
