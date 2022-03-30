package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/ot4i/ace-docker/common/logger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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

func Test_deployIntegrationFlowResources(t *testing.T) {
	t.Run("deployIntegrationFlowResources returns nil as both commands returned without error", func(t *testing.T) {
		oldcommandRunCmd := commandRunCmd
		defer func() { commandRunCmd = oldcommandRunCmd }()
		commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
			return "success", 0, nil
		}
		err := deployIntegrationFlowResources()
		assert.NoError(t, err)
	})

	t.Run("deployIntegrationFlowResources error running acecc-bar-gen.js", func(t *testing.T) {
		oldpackageBarFile := packageBarFile
		defer func() { packageBarFile = oldpackageBarFile }()
		packageBarFile = func() error {
			err := errors.New("Error running acecc-bar-gen.js")
			return err
		}

		oldcommandRunCmd := commandRunCmd
		defer func() { commandRunCmd = oldcommandRunCmd }()
		commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
			err := errors.New("Error running acecc-bar-gen.js")
			return "Error running acecc-bar-gen.js", 1, err
		}

		err := deployIntegrationFlowResources()
		assert.Error(t, err)
	})

	t.Run("deployIntegrationFlowResources error deploying BAR file - non zero return code", func(t *testing.T) {
		oldpackageBarFile := packageBarFile
		defer func() { packageBarFile = oldpackageBarFile }()
		packageBarFile = func() error {
			return nil
		}

		olddeployBarFile := deployBarFile
		defer func() { deployBarFile = olddeployBarFile }()
		deployBarFile = func() error {
			err := errors.New("Error deploying BAR file")
			return err
		}
		oldcommandRunCmd := commandRunCmd
		defer func() { commandRunCmd = oldcommandRunCmd }()
		commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
			err := errors.New("Error deploying BAR file")
			return "Error deploying BAR file", 1, err
		}

		err := deployIntegrationFlowResources()
		assert.Error(t, err)
	})
}

func Test_PackageBarFile(t *testing.T) {
	t.Run("Success running acecc-bar-gen.js", func(t *testing.T) {

		oldcommandRunCmd := commandRunCmd
		defer func() { commandRunCmd = oldcommandRunCmd }()
		commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
			return "Success", 0, nil
		}
		err := packageBarFile()
		assert.NoError(t, err)

	})
	t.Run("Error running acecc-bar-gen.js", func(t *testing.T) {

		oldcommandRunCmd := commandRunCmd
		defer func() { commandRunCmd = oldcommandRunCmd }()
		commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
			err := errors.New("Error running acecc-bar-gen.js")
			return "Error running acecc-bar-gen.js", 1, err
		}
		err := packageBarFile()
		assert.Error(t, err)
	})
}

func Test_DeployBarFile(t *testing.T) {
	t.Run("Success deploying BAR file", func(t *testing.T) {
		oldcommandRunCmd := commandRunCmd
		defer func() { commandRunCmd = oldcommandRunCmd }()
		commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
			return "Success", 0, nil
		}
		err := deployBarFile()
		assert.NoError(t, err)

	})
	t.Run("Error deploying BAR file", func(t *testing.T) {
		oldcommandRunCmd := commandRunCmd
		defer func() { commandRunCmd = oldcommandRunCmd }()
		commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
			err := errors.New("Error running deploying BAR file")
			return "Error running deploying BAR file", 1, err
		}
		err := deployBarFile()
		assert.Error(t, err)
	})

}

func Test_deployCSAPIFlows(t *testing.T) {

	t.Run("Env var not set", func(t *testing.T) {
		called := 0
		oldcommandRunCmd := commandRunCmd
		defer func() { commandRunCmd = oldcommandRunCmd }()
		commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
			called++
			return "Success", 0, nil
		}
		err := deployCSAPIFlows()
		assert.NoError(t, err)
		assert.Equal(t, 0, called)
	})

	t.Run("Env var not set", func(t *testing.T) {
		os.Setenv("CONNECTOR_SERVICE", "true")

		t.Run("Success deploying copyied files", func(t *testing.T) {
			var command []string
			oldcommandRunCmd := commandRunCmd
			defer func() { commandRunCmd = oldcommandRunCmd }()
			commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
				command = cmd.Args
				return "Success", 0, nil
			}
			err := deployCSAPIFlows()
			assert.NoError(t, err)
			assert.Contains(t, command, "/home/aceuser/deps/CSAPI")
			assert.Contains(t, command, "/home/aceuser/ace-server/run/CSAPI")

		})
		t.Run("Error copying files", func(t *testing.T) {
			oldcommandRunCmd := commandRunCmd
			defer func() { commandRunCmd = oldcommandRunCmd }()
			commandRunCmd = func(cmd *exec.Cmd) (string, int, error) {
				err := errors.New("Error running deploying BAR file")
				return "Error running deploying BAR file", 1, err
			}
			err := deployCSAPIFlows()
			assert.Error(t, err)
		})
		os.Unsetenv("CONNECTOR_SERVICE")
	})

}


func Test_forceflowbasicauthServerConfUpdate(t *testing.T) {
	t.Run("Golden path - Empty file gets populated with the right values", func(t *testing.T) {

		oldReadServerConfFile := readServerConfFile
		readServerConfFile = func() ([]byte, error) {
			return []byte{}, nil
		}
		oldWriteServerConfFileLocal := writeServerConfFile
		writeServerConfFile = func(content []byte) error {
			return nil
		}

		err := forceflowbasicauthServerConfUpdate()
		assert.NoError(t, err)
		readServerConfFile = oldReadServerConfFile
		writeServerConfFile = oldWriteServerConfFileLocal
	})

	t.Run("Existing value in server.conf.yaml", func(t *testing.T) {

		serverconfMap := make(map[interface{}]interface{})
		serverconfMap["forceServerHTTPSecurityProfile"] = "{DefaultPolicies}:SecProfLocal"
		serverconfYaml, _ := yaml.Marshal(&serverconfMap)

		oldReadServerConfFile := readServerConfFile
		readServerConfFile = func() ([]byte, error) {
			return serverconfYaml, nil
		}
		oldWriteServerConfFileLocal := writeServerConfFile
		writeServerConfFile = func(content []byte) error {
			return nil
		}

		err := forceflowbasicauthServerConfUpdate()
		assert.NoError(t, err)
		readServerConfFile = oldReadServerConfFile
		writeServerConfFile = oldWriteServerConfFileLocal
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

		err := forceflowbasicauthServerConfUpdate()
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

		err := forceflowbasicauthServerConfUpdate()
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

		err := forceflowbasicauthServerConfUpdate()
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
		err := forceflowbasicauthServerConfUpdate()
		assert.Error(t, err)
		assert.Equal(t, "Error writing server.conf.yaml", err.Error())

		yamlUnmarshal = oldYamlUnmarshal
		yamlMarshal = oldYamlMarshal
		writeServerConfFile = oldWriteServerConfFile
	})
}
