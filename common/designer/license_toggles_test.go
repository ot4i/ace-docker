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

package designer

import (
	"os"

	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertAppConnectLicenseToggleEnvironmentVariable(t *testing.T) {
	t.Run("When environment variable is empty", func(t *testing.T) {
		licenseToggles := map[string]bool{}
		err := ConvertAppConnectLicenseToggleEnvironmentVariable("", licenseToggles)
		require.NoError(t, err)
		require.Equal(t, 0, len(licenseToggles))
	})

	t.Run("When environment variable is invalid JSON", func(t *testing.T) {
		licenseToggles := map[string]bool{}
		err := ConvertAppConnectLicenseToggleEnvironmentVariable("{\"foo\":1,\"bar\":}", licenseToggles)
		require.Error(t, err)
		require.Equal(t, 0, len(licenseToggles))
	})

	t.Run("When environment variable is valid JSON", func(t *testing.T) {
		licenseToggles := map[string]bool{}
		err := ConvertAppConnectLicenseToggleEnvironmentVariable("{\"foo\":1,\"bar\":0}", licenseToggles)
		require.NoError(t, err)
		require.Equal(t, 2, len(licenseToggles))
		require.True(t, licenseToggles["foo"])
		require.False(t, licenseToggles["bar"])
	})
}

func TestGetLicenseTogglesFromEnvironmentVariables(t *testing.T) {
	os.Unsetenv("APP_CONNECT_LICENSE_TOGGLES")
	os.Unsetenv("APP_CONNECT_LICENSE_TOGGLES_OVERRIDE")

	t.Run("When only APP_CONNECT_LICENSE_TOGGLES is invalid JSON and fails to convert", func(t *testing.T) {
		os.Setenv("APP_CONNECT_LICENSE_TOGGLES", "{\"foo\":1,\"bar\":0")
		_, err := GetLicenseTogglesFromEnvironmentVariables()
		require.Error(t, err)
		os.Unsetenv("APP_CONNECT_LICENSE_TOGGLES")
	})

	t.Run("When only APP_CONNECT_LICENSE_TOGGLES_OVERRIDE is invalid JSON and fails to convert", func(t *testing.T) {
		os.Setenv("APP_CONNECT_LICENSE_TOGGLES_OVERRIDE", "{\"foo\":0")
		_, err := GetLicenseTogglesFromEnvironmentVariables()
		require.Error(t, err)
		os.Unsetenv("APP_CONNECT_LICENSE_TOGGLES_OVERRIDE")
	})

	t.Run("When neither APP_CONNECT_LICENSE_TOGGLES and APP_CONNECT_LICENSE_TOGGLES_OVERRIDE are empty", func(t *testing.T) {
		os.Setenv("APP_CONNECT_LICENSE_TOGGLES", "{\"foo\":1,\"bar\":0}")
		os.Setenv("APP_CONNECT_LICENSE_TOGGLES_OVERRIDE", "{\"bar\":1}")
		licenseToggles, err := GetLicenseTogglesFromEnvironmentVariables()
		require.NoError(t, err)
		require.True(t, licenseToggles["foo"])
		require.True(t, licenseToggles["bar"])
		os.Unsetenv("APP_CONNECT_LICENSE_TOGGLES")
		os.Unsetenv("APP_CONNECT_LICENSE_TOGGLES_OVERRIDE")
	})
}

func TestIsLicenseToggleEnabled (t *testing.T) {
	oldGlobalLicenseToggles := globalLicenseToggles
	globalLicenseToggles = map[string]bool{
		"foo": true,
		"bar": false,
	}
	
	require.True(t, isLicenseToggleEnabled("foo"))
	require.False(t, isLicenseToggleEnabled("bar"))
	require.True(t, isLicenseToggleEnabled("unknown"))
	globalLicenseToggles = oldGlobalLicenseToggles
}

func TestInitialiseLicenseToggles (t *testing.T) {
	licenseToggles := map[string]bool{
		"foo": true,
		"bar": false,
	}
	InitialiseLicenseToggles(licenseToggles)
	require.True(t, globalLicenseToggles["foo"])
	require.False(t, globalLicenseToggles["bar"])
}