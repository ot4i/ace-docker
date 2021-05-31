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

// Package name contains code to manage the queue manager name
package name

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

// sanitizeName removes any invalid characters from a queue manager name
func sanitizeName(name string) string {
	var re = regexp.MustCompile("[^a-zA-Z0-9._%/]")
	return re.ReplaceAllString(name, "")
}

// GetQueueManagerName resolves the queue manager name to use.  Resolved from
// either an environment variable, or the hostname.
func GetQueueManagerName() (string, error) {
	var name string
	var err error
	name, ok := os.LookupEnv("MQ_QMGR_NAME")
	if !ok || name == "" {
		name, err = os.Hostname()
		if err != nil {
			return "", err
		}
		name = sanitizeName(name)
	}
	return name, nil
}

// GetIntegrationServerName resolves the integration server naem to use.
// Resolved from either an environment variable, or the hostname.
func GetIntegrationServerName() (string, error) {
	var name string
	var err error
	name, ok := os.LookupEnv("ACE_SERVER_NAME")
	if !ok || name == "" {
		name, err = os.Hostname()
		if err != nil {
			return "", err
		}
		name = sanitizeName(name)
	}
	return name, nil
}

// GetIntegrationServerVaultKey resolves the integration server Vault key to use.
// If error, configuration was wrong.
// If empty and no error, there was nothing configured.
func GetIntegrationServerVaultKey() (string, error) {
	var path string
	var err error
	path, ok := os.LookupEnv("ACE_VAULT_KEY_FILE")
	if !ok || path == "" {
		return "", nil
	}
	stat, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if stat.IsDir() {
		return "", fmt.Errorf("vault key file is a directory: %s", path)
	}
	bytes, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return "", readErr
	}
	return string(bytes), nil
}
