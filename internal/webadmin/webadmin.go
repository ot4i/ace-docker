package webadmin

import (
	"crypto/rand"
	"crypto/sha512"
	b64 "encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ot4i/ace-docker/common/logger"
	"golang.org/x/crypto/pbkdf2"
	"gopkg.in/yaml.v2"
)

var (
	readFile                      = os.ReadFile
	mkdirAll                      = os.MkdirAll
	writeFile                     = os.WriteFile
	readDir                       = os.ReadDir
	processWebAdminUsers          = processWebAdminUsersLocal
	applyFileAuthOverrides        = applyFileAuthOverridesLocal
	outputFiles                   = outputFilesLocal
	unmarshal                     = yaml.Unmarshal
	marshal                       = yaml.Marshal
	readWebUsersTxt               = readWebUsersTxtLocal
	version                string = "12.0.0.0"
	homedir                string = "/home/aceuser/"
	webusersDir            string = "/home/aceuser/initial-config/webusers/"
)

func ConfigureWebAdminUsers(log logger.LoggerInterface) error {
	serverConfContent, err := readServerConfFile()
	if err != nil {
		log.Errorf("Error reading server.conf.yaml: %v", err)
		return err
	}

	webAdminUserInfo, err := processWebAdminUsers(log, webusersDir)
	if err != nil {
		log.Errorf("Error processing WebAdmin users: %v", err)
		return err
	}
	serverconfYaml, err := applyFileAuthOverrides(log, webAdminUserInfo, serverConfContent)
	if err != nil {
		log.Errorf("Error applying file auth overrides: %v", err)
		return err
	}
	err = writeServerConfFile(serverconfYaml)
	if err != nil {
		log.Errorf("Error writing server.conf.yaml: %v", err)
		return err
	}

	for webAdminUserName, webAdminPass := range webAdminUserInfo {
		m := map[string]string{
			"password": keyGen(webAdminPass),
			"role":     webAdminUserName,
			"version":  version,
		}
		err := outputFiles(log, m)
		if err != nil {
			log.Errorf("Error writing WebAdmin files: %v", err)
			return err
		}
	}
	return nil
}

func processWebAdminUsersLocal(log logger.LoggerInterface, dir string) (map[string]string, error) {
	userInfo := map[string]string{}

	fileList, err := readDir(dir)
	if err != nil {
		log.Errorf("Error reading directory: %v", err)
		return nil, err
	}

	for _, file := range fileList {
		if filepath.Ext(file.Name()) == ".txt" {
			username, password, err := readWebUsersTxt(log, dir+file.Name())
			if err != nil {
				log.Errorf("Error reading WebAdmin users.txt file: %v", err)
				return nil, err
			}
			userInfo[username] = password
		}
	}
	return userInfo, nil
}

func applyFileAuthOverridesLocal(log logger.LoggerInterface, webAdminUserInfo map[string]string, serverconfContent []byte) ([]byte, error) {
	serverconfMap := make(map[interface{}]interface{})
	err := unmarshal([]byte(serverconfContent), &serverconfMap)
	if err != nil {
		log.Errorf("Error unmarshalling server.conf.yaml content: %v", err)
		return nil, err
	}

	permissionsMap := map[string]string{
		"admin":    "read+:write+:execute+",
		"operator": "read+:write-:execute+",
		"editor":   "read+:write+:execute-",
		"audit":    "read+:write-:execute-",
		"viewer":   "read+:write-:execute-",
	}

	if serverconfMap["Security"] == nil {
		serverconfMap["Security"] = map[interface{}]interface{}{}
		security := serverconfMap["Security"].(map[interface{}]interface{})
		if security["DataPermissions"] == nil && security["Permissions"] == nil {
			security["DataPermissions"] = map[interface{}]interface{}{}
			dataPermissions := security["DataPermissions"].(map[interface{}]interface{})
			security["Permissions"] = map[interface{}]interface{}{}
			permissions := security["Permissions"].(map[interface{}]interface{})
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-admin"]; ok {
				dataPermissions["admin"] = permissionsMap["admin"]
				permissions["admin"] = permissionsMap["admin"]
			}
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-operator"]; ok {
				permissions["operator"] = permissionsMap["operator"]
			}
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-editor"]; ok {
				permissions["editor"] = permissionsMap["editor"]
			}
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-audit"]; ok {
				permissions["audit"] = permissionsMap["audit"]
			}
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-viewer"]; ok {
				permissions["viewer"] = permissionsMap["viewer"]
			}
		}
	} else {
		security := serverconfMap["Security"].(map[interface{}]interface{})
		if security["DataPermissions"] == nil && security["Permissions"] == nil {
			security["DataPermissions"] = map[interface{}]interface{}{}
			dataPermissions := security["DataPermissions"].(map[interface{}]interface{})
			security["Permissions"] = map[interface{}]interface{}{}
			permissions := security["Permissions"].(map[interface{}]interface{})
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-admin"]; ok {
				dataPermissions["admin"] = permissionsMap["admin"]
				permissions["admin"] = permissionsMap["admin"]
			}
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-operator"]; ok {
				permissions["operator"] = permissionsMap["operator"]
			}
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-editor"]; ok {
				permissions["editor"] = permissionsMap["editor"]
			}
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-audit"]; ok {
				permissions["audit"] = permissionsMap["audit"]
			}
			if _, ok := webAdminUserInfo["ibm-ace-dashboard-viewer"]; ok {
				permissions["viewer"] = permissionsMap["viewer"]
			}
		}
	}

	serverconfYaml, err := marshal(&serverconfMap)
	if err != nil {
		log.Errorf("Error marshalling server.conf.yaml overrides: %v", err)
		return nil, err
	}

	return serverconfYaml, nil

}

func keyGen(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	dk := pbkdf2.Key([]byte(password), salt, 65536, 64, sha512.New)
	return fmt.Sprintf("PBKDF2-SHA-512:%s:%s", b64EncodeString(salt), b64EncodeString(dk))
}

func readServerConfFile() ([]byte, error) {
	return readFile(homedir + "ace-server/overrides/server.conf.yaml")

}

func writeServerConfFile(content []byte) error {
	return writeFile(homedir+"ace-server/overrides/server.conf.yaml", content, 0644)
}

func outputFilesLocal(log logger.LoggerInterface, files map[string]string) error {
	dir := homedir + "ace-server/config/registry/integration_server/CurrentVersion/WebAdmin/user/"
	webadminDir := dir + files["role"]
	err := mkdirAll(webadminDir, 0755)
	if err != nil {
		log.Errorf("Error creating directories: %v", err)
		return err
	}

	for fileName, fileContent := range files {
		// The 'role' is populated from the users.txt files in initial-config e.g. admin-users.txt we need to trim this to the actual role which would be 'admin'
		fileContent = strings.TrimPrefix(fileContent, "ibm-ace-dashboard-")
		err := writeFile(webadminDir+"/"+fileName, []byte(fileContent), 0660)
		if err != nil {
			log.Errorf("Error writing files: %v %s", err, fileName)
			return err
		}

	}

	return nil
}

func b64EncodeString(data []byte) string {
	return b64.StdEncoding.EncodeToString(data)
}

func readWebUsersTxtLocal(log logger.LoggerInterface, filename string) (string, string, error) {
	out, err := readFile(filename)
	if err != nil {
		log.Errorf("Error reading WebAdmin users.txt file: %v", err)
		return "", "", err
	}

	credentials := strings.Fields(string(out))
	return credentials[0], credentials[1], nil
}
