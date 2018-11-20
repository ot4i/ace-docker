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

// Package command contains code to run external commands
package command

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/ot4i/ace-docker/internal/logger"
)

// A BackgroundCmd provides a handle to a backgrounded command and its completion state.
type BackgroundCmd struct {
	Cmd         *exec.Cmd
	ReturnCode  int
	ReturnError error
	Started     bool
	Finished    bool
	finishChan  chan bool
}

// RunCmd runs an OS command.  On Linux it waits for the command to
// complete and returns the exit status (return code).
// Do not use this function to run shell built-ins (like "cd"), because
// the error handling works differently
func RunCmd(cmd *exec.Cmd) (string, int, error) {
	// Run the command and wait for completion
	out, err := cmd.Output()
	if err != nil {
		// Assert that this is an ExitError
		exiterr, ok := err.(*exec.ExitError)
		// If the type assertion was correct, and we're on Linux
		if ok && runtime.GOOS == "linux" {
			status, ok := exiterr.Sys().(syscall.WaitStatus)
			if ok {
				return string(out), status.ExitStatus(), fmt.Errorf("%v: %v: %v", cmd.Path, err, string(exiterr.Stderr))
			}
		}
		return string(out), -1, err
	}
	return string(out), 0, nil
}

// Run runs an OS command.  On Linux it waits for the command to
// complete and returns the exit status (return code).
// Do not use this function to run shell built-ins (like "cd"), because
// the error handling works differently
func Run(name string, arg ...string) (string, int, error) {
	return RunCmd(exec.Command(name, arg...))
}

// RunCmdBackground runs an OS command.  On Linux it runs the command in
// the background, piping stdout/stderr to log.LogDirect.
// It returns a BackgroundCmd, containing the os/exec/cmd, an int channel
// and error channel that will have the return code and error written respectively,
// when the process exits
func RunCmdBackground(cmd *exec.Cmd, log *logger.Logger) BackgroundCmd {
	bgCmd := BackgroundCmd{cmd, 0, nil, false, false, make(chan bool)}

	stdoutChildPipe, err := cmd.StdoutPipe()
	if err != nil {
		bgCmd.ReturnCode = -1
		bgCmd.ReturnError = err
		return bgCmd
	}
	stderrChildPipe, err := cmd.StderrPipe()
	if err != nil {
		bgCmd.ReturnCode = -1
		bgCmd.ReturnError = err
		return bgCmd
	}

	// Write both stdout and stderr of the child to our logs
	stdoutScanner := bufio.NewScanner(stdoutChildPipe)
	stderrScanner := bufio.NewScanner(stderrChildPipe)

	go func() {
		for stdoutScanner.Scan() {
			log.LogDirect(stdoutScanner.Text() + "\n")
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			log.LogDirect(stderrScanner.Text() + "\n")
		}
	}()

	// Start the command in the background
	err = cmd.Start()
	if err != nil {
		bgCmd.ReturnCode = -1
		bgCmd.ReturnError = err
		return bgCmd
	}

	bgCmd.Started = true

	// Wait on the command and mark as finished when done
	go func() {
		err := cmd.Wait()
		bgCmd.ReturnCode = 0
		bgCmd.ReturnError = nil
		if err != nil {
			// Assert that this is an ExitError
			exiterr, ok := err.(*exec.ExitError)
			// If the type assertion was correct, and we're on Linux
			if ok && runtime.GOOS == "linux" {
				status, ok := exiterr.Sys().(syscall.WaitStatus)
				if ok {
					bgCmd.ReturnCode = status.ExitStatus()
					bgCmd.ReturnError = fmt.Errorf("%v: %v", cmd.Path, err)
				}
			}
			bgCmd.ReturnCode = -1
			bgCmd.ReturnError = err
		}
		bgCmd.Finished = true
		bgCmd.finishChan <- true
		close(bgCmd.finishChan)
	}()

	return bgCmd
}

// SigIntBackground sends the signal SIGINT to the backgrounded command wrapped by
// the BackgroundCommand struct
func SigIntBackground(bgCmd BackgroundCmd) {
	if bgCmd.Started && !bgCmd.Finished {
		bgCmd.Cmd.Process.Signal(syscall.SIGINT) // TODO returns an error
	}
}

// WaitOnBackground will wait until the process is marked as finished.
func WaitOnBackground(bgCmd BackgroundCmd) {
	if !bgCmd.Finished {
		bgCmd.Finished = <-bgCmd.finishChan
	}
}

// RunBackground runs an OS command.  On Linux it runs the command in the background
// and returns a channel to int that will have the return code written when
// the process exits
func RunBackground(name string, log *logger.Logger, arg ...string) BackgroundCmd {
	return RunCmdBackground(exec.Command(name, arg...), log)
}

// RunAsUser runs the specified command as the aceuser user.  If the current user
// already is the specified user then this calls through to RunCmd.
func RunAsUser(username string, name string, arg ...string) (string, int, error) {
	thisUser, err := user.Current()
	if err != nil {
		return "", 0, err
	}
	cmd := exec.Command(name, arg...)
	if strings.Compare(username, thisUser.Username) != 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		uid, gid, groups, err := LookupUser(username)
		if err != nil {
			return "", 0, err
		}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, Groups: groups}
	}
	return RunCmd(cmd)
}

// RunAsUserBackground runs the specified command as the aceuser user in the background.
// It returns a BackgroundCmd, containing the os/exec/cmd, an int channel
// and error channel that will have the return code and error written respectively,
// when the process exits
func RunAsUserBackground(username string, name string, log *logger.Logger, arg ...string) BackgroundCmd {
	cmd := exec.Command(name, arg...)
	thisUser, err := user.Current()
	if err != nil {
		bgCmd := BackgroundCmd{cmd, 0, nil, false, false, make(chan bool)}
		bgCmd.ReturnCode = -1
		bgCmd.ReturnError = err
		return bgCmd
	}
	if strings.Compare(username, thisUser.Username) != 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		uid, gid, groups, err := LookupUser(username)
		if err != nil {
			bgCmd := BackgroundCmd{cmd, 0, nil, false, false, make(chan bool)}
			bgCmd.ReturnCode = -1
			bgCmd.ReturnError = err
			return bgCmd
		}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, Groups: groups}
	}
	return RunCmdBackground(cmd, log)
}

// LookupAceuser looks up the UID, GID and supplementary groups of the aceuser user.  This is only
// allowed if the current user is "root"; otherwise this returns an error.
func LookupUser(username string) (uint32, uint32, []uint32, error) {
	thisUser, err := user.Current()
	if err != nil {
		return uint32(0), uint32(0), []uint32{}, err
	}
	if strings.Compare("root", thisUser.Username) != 0 {
		return uint32(0), uint32(0), []uint32{}, errors.New("Not permitted: the current user attempted to look up user but is not permitted.")
	}

	user, err := user.Lookup(username)
	if err != nil {
		return uint32(0), uint32(0), []uint32{}, err
	}
	userUID, err := strconv.Atoi(user.Uid)
	if err != nil {
		return uint32(0), uint32(0), []uint32{}, err
	}
	userGID, err := strconv.Atoi(user.Gid)
	if err != nil {
		return uint32(0), uint32(0), []uint32{}, err
	}
	userSupplementaryGIDStrings, err := user.GroupIds()
	if err != nil {
		return uint32(0), uint32(0), []uint32{}, err
	}
	var userSupplementaryGIDs []uint32
	for _, idString := range userSupplementaryGIDStrings {
		id, err := strconv.Atoi(idString)
		if err != nil {
			return uint32(0), uint32(0), []uint32{}, err
		}
		userSupplementaryGIDs = append(userSupplementaryGIDs, uint32(id))
	}

	return uint32(userUID), uint32(userGID), userSupplementaryGIDs, nil
}
