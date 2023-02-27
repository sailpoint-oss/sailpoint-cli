// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/sdk-output/v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

var (
	createTemplate = []byte(`{
  "attributes": {
    "substring": "admin_"
  },
  "type": "indexOf",
  "name": "Test Index Of Transform"
}`)

	path          = "test_data/"
	createFile    = "test_create.json"
	updateFile    = "test_update.json"
	testTransform sailpointsdk.Transform
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

func SaveTransform(filePath string) error {
	// Make sure to create the files if they dont exist
	file, err := os.OpenFile((filepath.Join(path, filePath)), os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	createString, err := json.Marshal(testTransform)
	if err != nil {
		return err
	}

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

	err := json.Unmarshal([]byte(createTemplate), &testTransform)
	if err != nil {
		t.Fatalf("Error unmarshalling template: %v", err)
	}

	testTransform.Name = randSeq(6)

	// Make sure the output dir exists first
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		t.Fatalf("Error unmarshalling template: %v", err)
	}

	err = SaveTransform(createFile)
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

	transformID := createBuffer.String()
	fmt.Println(transformID)

	testTransform.Attributes["substring"] = randSeq(24)
	testTransform.Id = &transformID

	err = SaveTransform(updateFile)
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
