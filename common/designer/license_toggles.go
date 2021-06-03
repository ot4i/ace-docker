/*
© Copyright IBM Corporation 2020

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

// Package designer contains code for the designer specific logic
package designer

import (
	"os"
	"encoding/json"
)

var globalLicenseToggles map[string]bool

// ConvertAppConnectLicenseToggleEnvironmentVariable converts the string representation of the environment variable
// into a <toggle>:<boolean> representation
func ConvertAppConnectLicenseToggleEnvironmentVariable (environmentVariable string, licenseToggles map[string]bool) error {
	var licenseTogglesValues map[string]int
	if environmentVariable != "" {
		err := json.Unmarshal([]byte(environmentVariable), &licenseTogglesValues)
		if err != nil {
			return err
		}
	}

	for licenseToggle, value := range licenseTogglesValues {
		if value == 1 {
			licenseToggles[licenseToggle] = true
		} else {
			licenseToggles[licenseToggle] = false
		}
	}
	return nil
}
// GetLicenseTogglesFromEnvironmentVariables reads the APP_CONNECT_LICENSE_TOGGLES and APP_CONNECT_LICENSE_TOGGLES_OVERRIDE environment variables
// it sets the license toggles to APP_CONNECT_LICENSE_TOGGLES and overrides the values using APP_CONNECT_LICENSE_TOGGLES_OVERRIDE
func GetLicenseTogglesFromEnvironmentVariables () (map[string]bool, error) {
	licenseToggles := map[string]bool{}

	err := ConvertAppConnectLicenseToggleEnvironmentVariable(os.Getenv("APP_CONNECT_LICENSE_TOGGLES"), licenseToggles)
	if err != nil {
		return nil, err
	}
	err = ConvertAppConnectLicenseToggleEnvironmentVariable(os.Getenv("APP_CONNECT_LICENSE_TOGGLES_OVERRIDE"), licenseToggles)
	if err != nil {
		return nil, err
	}
	return licenseToggles, nil
}

// isLicenseToggleEnabled checks if a toggle is enabled
// if not defined, then it's considered enabled
var isLicenseToggleEnabled = func (toggle string) bool {
	enabled, ok := globalLicenseToggles[toggle]
	return !ok || enabled
}

// InitialiseLicenseToggles initialises the globalLicenseToggles map
func InitialiseLicenseToggles(licenseToggles map[string]bool) {
	globalLicenseToggles = licenseToggles
}