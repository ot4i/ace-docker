package webadmin

import (
	b64 "encoding/base64"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ot4i/ace-docker/common/logger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// Allows us to read server.conf.yaml out into a struct which makes it easier check the values
type ServerConf struct {
	RestAdminListener struct {
		AuthorizationEnabled bool `yaml:"authorizationEnabled"`
	} `yaml:"RestAdminListener"`
	Security struct {
		LdapAuthorizeAttributeToRoleMap struct {
			RandomField string `yaml:"randomfield"`
		} `yaml:"LdapAuthorizeAttributeToRoleMap"`
		DataPermissions struct {
			Admin string `yaml:"admin"`
		} `yaml:"DataPermissions"`
		Permissions struct {
			Admin    string `yaml:"admin"`
			Audit    string `yaml:"audit"`
			Editor   string `yaml:"editor"`
			Operator string `yaml:"operator"`
			Viewer   string `yaml:"viewer"`
		} `yaml:"Permissions"`
	} `yaml:"Security"`
}
type ServerConfWithoutSecurity struct {
	RestAdminListener struct {
		AuthorizationEnabled bool `yaml:"authorizationEnabled"`
	} `yaml:"RestAdminListener"`
}

type ServerConfNoPermissions struct {
	RestAdminListener struct {
		AuthorizationEnabled bool `yaml:"authorizationEnabled"`
	} `yaml:"RestAdminListener"`
	Security struct {
		LdapAuthorizeAttributeToRoleMap struct {
			RandomField string `yaml:"randomfield"`
		} `yaml:"LdapAuthorizeAttributeToRoleMap"`
	} `yaml:"Security"`
}

func Test_ConfigureWebAdminUsers(t *testing.T) {
	log, _ := logger.NewLogger(os.Stdout, true, false, "testloger")
	oldReadFile := readFile
	oldProcessWebAdminUsers := processWebAdminUsers
	oldApplyFileAuthOverrides := applyFileAuthOverrides
	oldWriteFile := writeFile
	oldOutputFiles := outputFiles
	t.Run("Golden path - all functions are free from errors and ConfigureWebAdminUsers returns nil", func(t *testing.T) {
		readFile = func(name string) ([]byte, error) {
			return nil, nil
		}

		processWebAdminUsers = func(logger.LoggerInterface, string) (map[string]string, error) {
			usersMap := map[string]string{
				"ibm-ace-dashboard-admin": "12AB0C96-E155-43FA-BA03-BD93AA2166E0",
			}

			return usersMap, nil
		}

		applyFileAuthOverrides = func(log logger.LoggerInterface, webAdminUserInfo map[string]string, serverconfContent []byte) ([]byte, error) {

			return nil, nil
		}
		writeFile = func(name string, data []byte, perm os.FileMode) error {
			return nil
		}
		outputFiles = func(logger.LoggerInterface, map[string]string) error {
			return nil
		}
		err := ConfigureWebAdminUsers(log)
		assert.NoError(t, err)
		readFile = oldReadFile
		processWebAdminUsers = oldProcessWebAdminUsers
		applyFileAuthOverrides = oldApplyFileAuthOverrides
		writeFile = oldWriteFile
		outputFiles = oldOutputFiles

	})

	t.Run("readServerConfFile returns an error unable to read server.conf.yaml file", func(t *testing.T) {
		readFile = func(name string) ([]byte, error) {
			return nil, errors.New("Unable to read server.conf.yaml")
		}
		err := ConfigureWebAdminUsers(log)
		assert.Error(t, err)
		assert.Equal(t, "Unable to read server.conf.yaml", err.Error())
		readFile = oldReadFile

	})
	t.Run("processAdminUsers returns an error unable to process webadmin users", func(t *testing.T) {
		readFile = func(name string) ([]byte, error) {
			return nil, nil
		}

		processWebAdminUsers = func(log logger.LoggerInterface, dir string) (map[string]string, error) {
			return nil, errors.New("Unable to process web admin users")
		}

		err := ConfigureWebAdminUsers(log)
		assert.Error(t, err)
		assert.Equal(t, "Unable to process web admin users", err.Error())
		readFile = oldReadFile
		processWebAdminUsers = oldProcessWebAdminUsers

	})

	t.Run("applyFileAuthOverrides returns an error unable to apply file auth overrides", func(t *testing.T) {
		readFile = func(name string) ([]byte, error) {
			return nil, nil
		}

		processWebAdminUsers = func(log logger.LoggerInterface, dir string) (map[string]string, error) {
			return nil, nil
		}

		applyFileAuthOverrides = func(log logger.LoggerInterface, webAdminUserInfo map[string]string, serverconfContent []byte) ([]byte, error) {
			return nil, errors.New("Unable to apply file auth overrides")
		}

		err := ConfigureWebAdminUsers(log)
		assert.Error(t, err)
		assert.Equal(t, "Unable to apply file auth overrides", err.Error())
		readFile = oldReadFile
		processWebAdminUsers = oldProcessWebAdminUsers
		applyFileAuthOverrides = oldApplyFileAuthOverrides

	})

	t.Run("writeServerConfFile error writing server.conf.yaml overrides back into the file", func(t *testing.T) {
		readFile = func(name string) ([]byte, error) {
			return nil, nil
		}

		processWebAdminUsers = func(log logger.LoggerInterface, dir string) (map[string]string, error) {
			return nil, nil
		}

		applyFileAuthOverrides = func(log logger.LoggerInterface, webAdminUserInfo map[string]string, serverconfContent []byte) ([]byte, error) {
			return nil, nil
		}
		writeFile = func(name string, data []byte, perm os.FileMode) error {
			return errors.New("Error writing server.conf.yaml back after overrides")
		}
		err := ConfigureWebAdminUsers(log)
		assert.Error(t, err)
		assert.Equal(t, "Error writing server.conf.yaml back after overrides", err.Error())
		readFile = oldReadFile
		processWebAdminUsers = oldProcessWebAdminUsers
		applyFileAuthOverrides = oldApplyFileAuthOverrides
		writeFile = oldWriteFile

	})

	t.Run("writeServerConfFile error writing server.conf.yaml overrides back into the file", func(t *testing.T) {
		readFile = func(name string) ([]byte, error) {
			return nil, nil
		}

		processWebAdminUsers = func(logger.LoggerInterface, string) (map[string]string, error) {
			usersMap := map[string]string{
				"ibm-ace-dashboard-admin": "12AB0C96-E155-43FA-BA03-BD93AA2166E0",
			}

			return usersMap, nil
		}

		applyFileAuthOverrides = func(log logger.LoggerInterface, webAdminUserInfo map[string]string, serverconfContent []byte) ([]byte, error) {

			return nil, nil
		}
		writeFile = func(name string, data []byte, perm os.FileMode) error {
			return nil
		}
		outputFiles = func(logger.LoggerInterface, map[string]string) error {
			return errors.New("Error outputting files during password generation")
		}
		err := ConfigureWebAdminUsers(log)
		assert.Error(t, err)
		assert.Equal(t, "Error outputting files during password generation", err.Error())
		readFile = oldReadFile
		processWebAdminUsers = oldProcessWebAdminUsers
		applyFileAuthOverrides = oldApplyFileAuthOverrides
		writeFile = oldWriteFile
		outputFiles = oldOutputFiles
	})
}

func Test_processWebAdminUsers(t *testing.T) {
	log, _ := logger.NewLogger(os.Stdout, true, false, "testloger")
	t.Run("readDir returns error reading directory", func(t *testing.T) {
		oldReadDir := readDir
		readDir = func(name string) ([]os.DirEntry, error) {
			return nil, errors.New("Error reading directory")
		}
		_, err := processWebAdminUsers(log, "dir")
		assert.Error(t, err)
		assert.Equal(t, "Error reading directory", err.Error())

		readDir = oldReadDir
	})
	t.Run("Golden path - processWebAdminUsers loops over fileList and falls readWebUsersTxt for each .txt file ", func(t *testing.T) {
		webAdminUsers, err := processWebAdminUsers(log, "testdata/initial-config/webusers/")
		assert.NoError(t, err)
		assert.Equal(t, "1758F07A-8BEF-448C-B020-C25946AF3E94", webAdminUsers["ibm-ace-dashboard-admin"])
		assert.Equal(t, "68FE7808-8EC2-4395-97D0-A776D2A61912", webAdminUsers["ibm-ace-dashboard-operator"])
		assert.Equal(t, "28DBC34B-C0FD-44BF-8100-99DB686B6DB2", webAdminUsers["ibm-ace-dashboard-editor"])
		assert.Equal(t, "929064C2-0017-4B34-A883-219A4D1AC944", webAdminUsers["ibm-ace-dashboard-audit"])
		assert.Equal(t, "EF086556-74B8-4FB0-ACF8-CC59E1F3DB5F", webAdminUsers["ibm-ace-dashboard-viewer"])
	})

	t.Run("readWebUsersTxt fails to read files", func(t *testing.T) {
		oldReadWebUsersTxt := readWebUsersTxt
		readWebUsersTxt = func(logger logger.LoggerInterface, filename string) (string, string, error) {
			return "", "", errors.New("Error reading WebAdmin users txt file")
		}
		_, err := processWebAdminUsers(log, "testdata/initial-config/webusers")
		assert.Error(t, err)
		assert.Equal(t, "Error reading WebAdmin users txt file", err.Error())

		readWebUsersTxt = oldReadWebUsersTxt
	})

}

func Test_applyFileAuthOverrides(t *testing.T) {
	log, _ := logger.NewLogger(os.Stdout, true, false, "testloger")
	usersMap := map[string]string{
		"ibm-ace-dashboard-admin":    "12AB0C96-E155-43FA-BA03-BD93AA2166E0",
		"ibm-ace-dashboard-operator": "12AB0C96-E155-43FA-BA03-BD93AA2166E0",
		"ibm-ace-dashboard-editor":   "12AB0C96-E155-43FA-BA03-BD93AA2166E0",
		"ibm-ace-dashboard-audit":    "12AB0C96-E155-43FA-BA03-BD93AA2166E0",
		"ibm-ace-dashboard-viewer":   "12AB0C96-E155-43FA-BA03-BD93AA2166E0",
	}
	t.Run("Golden path - server.conf.yaml is populated as expected", func(t *testing.T) {
		// Pass in server.conf.yaml with some fields populated to prove we don't remove existing overrides
		servConf := &ServerConfWithoutSecurity{}
		servConf.RestAdminListener.AuthorizationEnabled = true
		servConfByte, err := yaml.Marshal(servConf)
		assert.NoError(t, err)

		serverConfContent, err := applyFileAuthOverrides(log, usersMap, servConfByte)
		assert.NoError(t, err)

		serverconfMap := make(map[interface{}]interface{})
		err = yaml.Unmarshal(serverConfContent, &serverconfMap)
		assert.NoError(t, err)
		// This struct has the security tab so that it can parse all the information out to checked in assertions
		var serverConfWithSecurity ServerConf
		err = yaml.Unmarshal(serverConfContent, &serverConfWithSecurity)
		if err != nil {
			t.Log(err)
			t.Fail()

		}

		assert.Equal(t, "read+:write+:execute+", serverConfWithSecurity.Security.DataPermissions.Admin)
		assert.Equal(t, "read+:write+:execute+", serverConfWithSecurity.Security.Permissions.Admin)

		assert.Equal(t, "read+:write-:execute+", serverConfWithSecurity.Security.Permissions.Operator)
		assert.Equal(t, "read+:write+:execute-", serverConfWithSecurity.Security.Permissions.Editor)
		assert.Equal(t, "read+:write-:execute-", serverConfWithSecurity.Security.Permissions.Audit)
		assert.Equal(t, "read+:write-:execute-", serverConfWithSecurity.Security.Permissions.Viewer)

	})
	t.Run("server.conf.yaml has a Security entry but no entry for DataPermissions or Permissions - to prove we still change permissions if security exists in yaml", func(t *testing.T) {

		servConf := &ServerConfNoPermissions{}
		servConf.Security.LdapAuthorizeAttributeToRoleMap.RandomField = "randomstring"
		servConfByte, err := yaml.Marshal(servConf)
		assert.NoError(t, err)

		serverConfContent, err := applyFileAuthOverrides(log, usersMap, servConfByte)
		assert.NoError(t, err)

		serverconfMap := make(map[interface{}]interface{})
		err = yaml.Unmarshal(serverConfContent, &serverconfMap)
		assert.NoError(t, err)

		// If the Permissions or DataPermissions do not exist we create them and therefore the below struct is to parse them into to make the `assert.Equal` checks easy
		var serverConfWithSecurity ServerConf
		err = yaml.Unmarshal(serverConfContent, &serverConfWithSecurity)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		assert.Equal(t, "read+:write+:execute+", serverConfWithSecurity.Security.DataPermissions.Admin)
		assert.Equal(t, "read+:write+:execute+", serverConfWithSecurity.Security.Permissions.Admin)

		assert.Equal(t, "read+:write-:execute+", serverConfWithSecurity.Security.Permissions.Operator)
		assert.Equal(t, "read+:write+:execute-", serverConfWithSecurity.Security.Permissions.Editor)
		assert.Equal(t, "read+:write-:execute-", serverConfWithSecurity.Security.Permissions.Audit)
		assert.Equal(t, "read+:write-:execute-", serverConfWithSecurity.Security.Permissions.Viewer)

	})

	t.Run("Unable to unmarhsall server conf into map for parsing", func(t *testing.T) {
		oldUnmarshal := unmarshal
		unmarshal = func(in []byte, out interface{}) (err error) {
			return errors.New("Unable to unmarshall server conf")
		}

		_, err := applyFileAuthOverrides(log, usersMap, []byte{})
		assert.Error(t, err)
		assert.Equal(t, "Unable to unmarshall server conf", err.Error())
		unmarshal = oldUnmarshal
	})

	t.Run("Unable to marshall server conf after processing", func(t *testing.T) {
		oldMarshal := marshal
		marshal = func(in interface{}) (out []byte, err error) {
			return nil, errors.New("Unable to marshall server conf")
		}
		_, err := applyFileAuthOverrides(log, usersMap, []byte{})
		assert.Error(t, err)
		assert.Equal(t, "Unable to marshall server conf", err.Error())
		marshal = oldMarshal
	})
}
func Test_KeyGen(t *testing.T) {
	// Result is of the format ALGORITHM:SALT:ENCRYPTED-PASSWORD
	result := keyGen("afc6dd77-ee58-4a51-8ecd-26f55e2ce2fb")
	splitResult := strings.Split(result, ":")

	// Decode salt
	decodedString, err := b64.StdEncoding.DecodeString(splitResult[1])
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	// Decode password
	decodedPasswordString, err := b64.StdEncoding.DecodeString(splitResult[2])
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, "PBKDF2-SHA-512", splitResult[0])
	assert.Equal(t, 16, len(decodedString))
	assert.Equal(t, 64, len(decodedPasswordString))
}
func Test_readServerConfFile(t *testing.T) {
	readFile = func(name string) ([]byte, error) {
		serverConfContent := []byte("this is a fake server.conf.yaml file")
		return serverConfContent, nil
	}
	serverConfContent, err := readServerConfFile()
	assert.NoError(t, err)
	assert.Equal(t, "this is a fake server.conf.yaml file", string(serverConfContent))
}

func Test_writeServerConfFile(t *testing.T) {
	writeFile = func(name string, data []byte, perm os.FileMode) error {
		return nil
	}
	serverConfContent := []byte{}
	err := writeServerConfFile(serverConfContent)
	assert.NoError(t, err)
}

func Test_outputFiles(t *testing.T) {
	log, _ := logger.NewLogger(os.Stdout, true, false, "testloger")
	m := map[string]string{
		"password": "password1234",
		"role":     "ibm-ace-dashboard-admin",
		"version":  "12.0.0.0",
	}
	t.Run("Golden path scenario - Outputting all files is successful and we get a nil error", func(t *testing.T) {
		oldmkdirAll := mkdirAll
		oldwriteFile := writeFile
		mkdirAll = func(path string, perm os.FileMode) error {
			return nil
		}
		writeFile = func(name string, data []byte, perm os.FileMode) error {
			/*
				This UT will fail if the code to trim 'ibm-ace-dashboard-' from the contents of the 'role' file gets removed.
				This contents gets read in dfrom users.txt file e.g. admin-users.txt with 'ibm-ace-dashboard-<ROLE> PASSWORD' as the format.
				Example - 'ibm-ace-dashboard-admin 08FDD35A-6EA0-4D48-A87D-E6373D414824'
				We need to trim the 'ibm-ace-dashboard-admin' down to 'admin' as that is the role that is used in the server.conf.yaml overrides.
			*/
			if strings.Contains(name, "role") {
				if strings.Contains(string(data), "ibm-ace-dashboard-") {
					t.Log("writeFile should be called with only the role e.g. 'admin' and not 'ibm-ace-dashboard-admin'")
					t.Fail()
				}
			}
			return nil
		}
		err := outputFiles(log, m)
		assert.NoError(t, err)
		mkdirAll = oldmkdirAll
		writeFile = oldwriteFile
	})

	t.Run("mkdirAll fails to create the directories for WebAdmin users", func(t *testing.T) {
		oldmkdirAll := mkdirAll
		mkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("mkdirAll fails to create WebAdmin users text")
		}
		outputFiles(log, m)
		mkdirAll = oldmkdirAll
	})

	t.Run("Writing the file to disk fails", func(t *testing.T) {
		oldmkdirAll := mkdirAll
		mkdirAll = func(path string, perm os.FileMode) error {
			return nil
		}
		oldwriteFile := writeFile
		writeFile = func(name string, data []byte, perm os.FileMode) error {
			return errors.New("Unable to write files to disk")
		}
		err := outputFiles(log, m)
		assert.Error(t, err)
		mkdirAll = oldmkdirAll
		writeFile = oldwriteFile
	})

}
func Test_b64EncodeString(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Base64 the following text - hello",
			args: args{data: []byte("hello")},
			want: "aGVsbG8=",
		},
		{
			name: "Base64 the following text - randomtext",
			args: args{data: []byte("randomtext")},
			want: "cmFuZG9tdGV4dA==",
		},
		{
			name: "Base64 the following text - afc6dd77-ee58-4a51-8ecd-26f55e2ce2fb",
			args: args{data: []byte("afc6dd77-ee58-4a51-8ecd-26f55e2ce2fb")},
			want: "YWZjNmRkNzctZWU1OC00YTUxLThlY2QtMjZmNTVlMmNlMmZi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b64EncodeString(tt.args.data); got != tt.want {
				t.Errorf("b64EncodeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readWebUsersTxt(t *testing.T) {
	log, _ := logger.NewLogger(os.Stdout, true, false, "testloger")
	t.Run("readWebUsersTxt success scenario - returns the username and password with nil error", func(t *testing.T) {
		readFile = func(name string) ([]byte, error) {
			return []byte("ibm-ace-dashboard-admin afc6dd77-ee58-4a51-8ecd-26f55e2ce2f"), nil
		}

		username, password, err := readWebUsersTxt(log, "admin-users.txt")
		assert.NoError(t, err)
		assert.Equal(t, "ibm-ace-dashboard-admin", username)
		assert.Equal(t, "afc6dd77-ee58-4a51-8ecd-26f55e2ce2f", password)

	})
	t.Run("readWebUsersTxt failure scenario - returns empty username and password with error", func(t *testing.T) {
		readFile = func(name string) ([]byte, error) {
			return []byte{}, errors.New("Error reading file")
		}

		username, password, err := readWebUsersTxt(log, "admin-users.txt")
		assert.Equal(t, "", username)
		assert.Equal(t, "", password)
		assert.Error(t, err)
	})
}
