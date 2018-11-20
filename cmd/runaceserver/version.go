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
	"errors"
	"regexp"
	"strings"

	"github.com/ot4i/ace-docker/internal/command"
)

var (
	// ImageCreated is the date the image was built
	ImageCreated = "Not specified"
	// ImageRevision is the source control revision identifier
	ImageRevision = "Not specified"
	// ImageSource is the URL to get source code for building the image
	ImageSource = "Not specified"
)

func logDateStamp() {
	log.Printf("Image created: %v", ImageCreated)
}

func logGitRepo() {
	log.Printf("Image revision: %v", ImageRevision)
}

func logGitCommit() {
	log.Printf("Image source: %v", ImageSource)
}

func extractVersion(mqsiversion string) (string, error) {
	versionRegex := regexp.MustCompile("(?sm).+BIP8996I: Version:\\s+(.*?)\\s?[\\r\\n].*")
	version := versionRegex.FindStringSubmatch(mqsiversion)
	if len(version) < 2 {
		return "", errors.New("Failed to find version")
	}
	return version[1], nil
}

func extractLevel(mqsiversion string) (string, error) {
	levelRegex := regexp.MustCompile("(?sm).+BIP8998I: Code Level:\\s+(.*?)\\s?[\\r\\n].*")
	level := levelRegex.FindStringSubmatch(mqsiversion)
	if len(level) < 2 {
		return "", errors.New("Failed to find level")
	}
	return level[1], nil
}

func extractBuildType(mqsiversion string) (string, error) {
	buildTypeRegex := regexp.MustCompile("(?sm).+BIP8999I: Build Type:\\s+(.*?)\\s[\\r\\n].*")
	buildType := buildTypeRegex.FindStringSubmatch(mqsiversion)
	if len(buildType) < 2 {
		return "", errors.New("Failed to find build type")
	}
	return buildType[1], nil
}

func logACEVersion() {
	out, _, err := command.Run("ace_mqsicommand.sh", "service", "-v")
	if err != nil {
		log.Printf("Error calling mqsiservice. Output: %v Error: %v", strings.TrimSuffix(string(out), "\n"), err)
	}

	version, err := extractVersion(out)
	if err != nil {
		log.Printf("Error getting ACE version: Output: %v Error: %v", strings.TrimSuffix(string(out), "\n"), err)
	}

	level, err := extractLevel(out)
	if err != nil {
		log.Printf("Error getting ACE level: Output: %v Error: %v", strings.TrimSuffix(string(out), "\n"), err)
	}

	buildType, err := extractBuildType(out)
	if err != nil {
		log.Printf("Error getting ACE build type: Output: %v Error: %v", strings.TrimSuffix(string(out), "\n"), err)
	}

	log.Printf("ACE version: %v", version)
	log.Printf("ACE level: %v", level)
	log.Printf("ACE build type: %v", buildType)
}

func logVersionInfo() {
	logDateStamp()
	logGitRepo()
	logGitCommit()
	logACEVersion()
}
