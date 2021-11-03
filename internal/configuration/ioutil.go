package configuration

import (
	"io"
	"io/ioutil"
	"os"
)

var ioutilReadFile = ioutil.ReadFile
var osMkdirAll = os.MkdirAll
var osMkdir = os.Mkdir
var osCreate = os.Create
var ioutilWriteFile = ioutil.WriteFile
var osOpenFile = os.OpenFile
var ioCopy = io.Copy
var osStat = os.Stat
var osIsNotExist = os.IsNotExist
var osRemoveAll = os.RemoveAll

var internalAppendFile = func(fileName string, fileContent []byte, filePerm os.FileMode) error {

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerm)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(fileContent)

	if err == nil {
		err = file.Sync()
	}

	if err != nil {
		return err
	}

	return nil
}
