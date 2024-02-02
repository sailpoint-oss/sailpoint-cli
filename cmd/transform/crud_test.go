// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package transform

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"

	PATH "path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	v3 "github.com/sailpoint-oss/golang-sdk/v2/api_v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

var (
	createTemplate = []byte(`{
  "attributes": {
    "substring": "admin_"
  },
  "type": "indexOf",
  "name": "Test Index Of Transform"
}`)

	path       = "test_data"
	createFile = "test_create.json"
	updateFile = "test_update.json"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func SaveTransform(fileName string, transform map[string]interface{}) error {
	file, err := os.OpenFile((PATH.Join(path, fileName)), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	createString, err := json.Marshal(transform)
	if err != nil {
		return err
	}

	log.Info("Saving Transform", "Indented", string(createString))

	_, err = file.Write(createString)
	if err != nil {
		return err
	}

	err = config.InitConfig()
	if err != nil {
		return err
	}

	return nil
}
func TestNewCRUDCmd(t *testing.T) {

	var transform v3.Transform

	err := json.Unmarshal([]byte(createTemplate), &transform)
	if err != nil {
		t.Fatalf("Error unmarshalling template: %v", err)
	}

	transformName := randSeq(16)

	createTransform := make(map[string]interface{})
	createTransform["name"] = transformName
	createTransform["type"] = transform.Type
	createTransform["attributes"] = transform.Attributes

	t.Log(util.PrettyPrint(createTransform))

	// Make sure the output dir exists first
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		t.Fatalf("Error Creating Folders: %v", err)
	}

	err = SaveTransform(createFile, createTransform)
	if err != nil {
		t.Fatalf("Unable to save test data: %v", err)
	}

	// ctrl := gomock.NewController(t)
	// defer ctrl.Finish()

	createCMD := newCreateCommand()

	createBuffer := new(bytes.Buffer)
	createCMD.SetOut(createBuffer)
	createCMD.Flags().Set("file", PATH.Join(path, createFile))

	err = createCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewCreateCmd: Unable to execute the command successfully: %v", err)
	}

	transformID := string(createBuffer.String())
	t.Log(transformID)

	Attributes := make(map[string]string)
	Attributes["substring"] = randSeq(24)

	updateTransform := make(map[string]interface{})
	updateTransform["attributes"] = Attributes
	updateTransform["name"] = transformName
	updateTransform["type"] = transform.Type
	updateTransform["id"] = transformID
	updateTransform["internal"] = false

	t.Log(util.PrettyPrint(updateTransform))

	err = SaveTransform(updateFile, updateTransform)
	if err != nil {
		t.Fatalf("Unable to save test data: %v", err)
	}

	cmd := newUpdateCommand()

	cmd.Flags().Set("file", PATH.Join(path, updateFile))

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("error execute cmd: %v", err)
	}

	deleteCMD := newDeleteCommand()

	deleteBuffer := new(bytes.Buffer)
	deleteCMD.SetOut(deleteBuffer)
	deleteCMD.SetArgs([]string{transformID})

	err = deleteCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewDeleteCmd: Unable to execute the command successfully: %v", err)
	}
}
