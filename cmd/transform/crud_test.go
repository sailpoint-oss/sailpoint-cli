// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package transform

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	PATH "path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/golang/mock/gomock"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/v3"
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

func SaveTransform(filePath string, transform map[string]interface{}) error {
	// Make sure to create the files if they dont exist
	file, err := os.OpenFile((PATH.Join(path, filePath)), os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	createString, err := json.MarshalIndent(transform, "", " ")
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

	var transform sailpointsdk.Transform

	err := json.Unmarshal([]byte(createTemplate), &transform)
	if err != nil {
		t.Fatalf("Error unmarshalling template: %v", err)
	}

	transformName := randSeq(6)

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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	createCMD := newCreateCmd()

	createBuffer := new(bytes.Buffer)
	createCMD.SetOut(createBuffer)
	createCMD.Flags().Set("file", filepath.Join(path, createFile))

	err = createCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewCreateCmd: Unable to execute the command successfully: %v", err)
	}

	transformID := string(createBuffer.String())
	t.Log(transformID)

	Attributes := make(map[string]interface{})
	Attributes["substring"] = randSeq(24)

	updateTransform := make(map[string]interface{})
	updateTransform["attributes"] = Attributes
	updateTransform["name"] = transform.Name
	updateTransform["type"] = transform.Type
	updateTransform["id"] = transformID

	t.Log(util.PrettyPrint(updateTransform))

	err = SaveTransform(updateFile, updateTransform)
	if err != nil {
		t.Fatalf("Unable to save test data: %v", err)
	}

	cmd := newUpdateCmd()

	cmd.Flags().Set("file", filepath.Join(path, updateFile))

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("error execute cmd: %v", err)
	}

	deleteCMD := newDeleteCmd()

	deleteBuffer := new(bytes.Buffer)
	deleteCMD.SetOut(deleteBuffer)
	deleteCMD.SetArgs([]string{transformID})

	err = deleteCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewDeleteCmd: Unable to execute the command successfully: %v", err)
	}
}
