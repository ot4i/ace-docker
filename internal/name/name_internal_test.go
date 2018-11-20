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
package name

import (
	"testing"
)

var sanitizeTests = []struct {
	in  string
	out string
}{
	{"foo", "foo"},
	{"foo-0", "foo0"},
	{"foo-", "foo"},
	{"-foo", "foo"},
	{"foo_0", "foo_0"},
}

func TestSanitizeName(t *testing.T) {
	for _, table := range sanitizeTests {
		s := sanitizeName(table.in)
		if s != table.out {
			t.Errorf("sanitizeName(%v) - expected %v, got %v", table.in, table.out, s)
		}
	}
}
