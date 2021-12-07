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
	"os"
	"regexp"
)

// sanitizeName removes any invalid characters from a queue manager name
func sanitizeName(name string) string {
	var re = regexp.MustCompile("[^a-zA-Z0-9._%/]")
	return re.ReplaceAllString(name, "")
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
