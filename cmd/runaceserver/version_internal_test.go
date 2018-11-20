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
package main

import (
	"testing"
)

var versionTests = []struct {
	in  string
	out string
}{
	{"BIPmsgs  en_US\n  Console CCSID=1208, ICU CCSID=1208\n  Default codepage=UTF-8, in ascii=UTF-8\n  JAVA console codepage name=UTF-8\n\nBIP8996I: Version:    11001 \nBIP8997I: Product:    IBM App Connect Enterprise \nBIP8998I: Code Level: S000-L180810.12854 \nBIP8999I: Build Type: Production, 64 bit, amd64_linux_2 \n\nBIP8974I: Component: DFDL-C, Build ID: 20180710-2330, Version: 1.1.2.0 (1.1.2.0), Platform: linux_x86 64-bit, Type: production \n\nBIP8071I: Successful command completion.", "11001"},
}

var levelTests = []struct {
	in  string
	out string
}{
	{"BIPmsgs  en_US\n  Console CCSID=1208, ICU CCSID=1208\n  Default codepage=UTF-8, in ascii=UTF-8\n  JAVA console codepage name=UTF-8\n\nBIP8996I: Version:    11001 \nBIP8997I: Product:    IBM App Connect Enterprise \nBIP8998I: Code Level: S000-L180810.12854 \nBIP8999I: Build Type: Production, 64 bit, amd64_linux_2 \n\nBIP8974I: Component: DFDL-C, Build ID: 20180710-2330, Version: 1.1.2.0 (1.1.2.0), Platform: linux_x86 64-bit, Type: production \n\nBIP8071I: Successful command completion.", "S000-L180810.12854"},
	{"BIPmsgs  en_US\n  Console CCSID=1208, ICU CCSID=1208\n  Default codepage=UTF-8, in ascii=UTF-8\n  JAVA console codepage name=UTF-8\n\nBIP8996I: Version:    11001 \nBIP8997I: Product:    IBM App Connect Enterprise \nBIP8998I: Code Level: S000-L180810.12854 Suffix \nBIP8999I: Build Type: Production, 64 bit, amd64_linux_2 \n\nBIP8974I: Component: DFDL-C, Build ID: 20180710-2330, Version: 1.1.2.0 (1.1.2.0), Platform: linux_x86 64-bit, Type: production \n\nBIP8071I: Successful command completion.", "S000-L180810.12854 Suffix"},
}

var buildTypeTests = []struct {
	in  string
	out string
}{
	{"BIPmsgs  en_US\n  Console CCSID=1208, ICU CCSID=1208\n  Default codepage=UTF-8, in ascii=UTF-8\n  JAVA console codepage name=UTF-8\n\nBIP8996I: Version:    11001 \nBIP8997I: Product:    IBM App Connect Enterprise \nBIP8998I: Code Level: S000-L180810.12854 \nBIP8999I: Build Type: Production, 64 bit, amd64_linux_2 \n\nBIP8974I: Component: DFDL-C, Build ID: 20180710-2330, Version: 1.1.2.0 (1.1.2.0), Platform: linux_x86 64-bit, Type: production \n\nBIP8071I: Successful command completion.", "Production, 64 bit, amd64_linux_2"},
}

func TestExtractVersion(t *testing.T) {
	for _, table := range versionTests {
		out, err := extractVersion(table.in)
		if err != nil {
			t.Errorf("extractVersion(%v) - unexpected error %v", table.in, err)
		}
		if out != table.out {
			t.Errorf("extractVersion(%v) - expected %v, got %v", table.in, table.out, out)
		}
	}
	_, err := extractVersion("xxx")
	if err == nil {
		t.Error("extractVersion(xxx) - expected an error but didn't get one")
	}
}

func TestExtractLevel(t *testing.T) {
	for _, table := range levelTests {
		out, err := extractLevel(table.in)
		if err != nil {
			t.Errorf("extractLevel(%v) - unexpected error %v", table.in, err)
		}
		if out != table.out {
			t.Errorf("extractLevel(%v) - expected %v, got %v", table.in, table.out, out)
		}
	}
	_, err := extractLevel("xxx")
	if err == nil {
		t.Error("extractLevel(xxx) - expected an error but didn't get one")
	}
}

func TestExtractBuildType(t *testing.T) {
	for _, table := range buildTypeTests {
		out, err := extractBuildType(table.in)
		if err != nil {
			t.Errorf("extractBuildType(%v) - unexpected error %v", table.in, err)
		}
		if out != table.out {
			t.Errorf("extractBuildType(%v) - expected %v, got %v", table.in, table.out, out)
		}
	}
	_, err := extractBuildType("xxx")
	if err == nil {
		t.Error("extractBuildType(xxx) - expected an error but didn't get one")
	}
}
