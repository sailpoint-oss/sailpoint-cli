// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/log"
)

var (
	createTemplate = []byte(`{
  "name": "Timestamp To Date",
  "type": "dateFormat",
  "attributes": {
    "sourceName": "Workday",
    "attributeName": "DEPARTMENT",
    "accountSortAttribute": "created",
    "accountSortDescending": false,
    "accountReturnFirstLink": false,
    "accountFilter": "!(nativeIdentity.startsWith(\"*DELETED*\"))",
    "accountPropertyFilter": "(groups.containsAll({'Admin'}) || location == 'Austin')",
    "requiresPeriodicRefresh": false,
    "input": {
      "type": "accountAttribute",
      "attributes": {
        "attributeName": "first_name",
        "sourceName": "Source"
      }
    }
  }
}`)

	path       = "test_data"
	createFile = "test_create.json"
	updateFile = "test_update.json"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func SaveTransform(fileName string, transform map[string]interface{}) error {
	file, err := os.OpenFile((filepath.Join(path, fileName)), os.O_RDWR|os.O_CREATE, 0666)
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

	return nil
}

func TestNewCRUDCmd(t *testing.T) {
	var transform map[string]interface{}

	err := json.Unmarshal([]byte(createTemplate), &transform)
	if err != nil {
		t.Fatalf("Error unmarshalling template: %v", err)
	}

	transformName := randSeq(16)
	transform["name"] = transformName

	// Make sure the output dir exists first
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		t.Fatalf("Error Creating Folders: %v", err)
	}

	err = SaveTransform(createFile, transform)
	if err != nil {
		t.Fatalf("Unable to save test data: %v", err)
	}

	// Create transform
	createCMD := newPostCmd()
	createBuffer := new(bytes.Buffer)
	createCMD.SetOut(createBuffer)
	createCMD.SetArgs([]string{"/v2024/transforms"})
	createCMD.Flags().Set("body-file", filepath.Join(path, createFile))
	createCMD.Flags().Set("jsonpath", "$.id")

	err = createCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewCreateCmd: Unable to execute the command successfully: %v", err)
	}

	// Read the output
	responseBytes, err := io.ReadAll(createBuffer)
	if err != nil {
		t.Fatalf("Error reading stdout: %v", err)
	}
	transformID := string(responseBytes)
	log.Info("Transform ID", "ID", transformID)

	// Validate the transform was created by getting it
	getCMD := newGetCmd()
	getBuffer := new(bytes.Buffer)
	getCMD.SetOut(getBuffer)
	getCMD.SetArgs([]string{"/v2024/transforms/" + transformID})
	getCMD.Flags().Set("jsonpath", "$.name")

	err = getCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewGetCmd: Unable to execute the command successfully: %v", err)
	}

	// Read the output
	responseBytes, err = io.ReadAll(getBuffer)
	if err != nil {
		t.Fatalf("Error reading stdout: %v", err)
	}
	retrievedName := string(responseBytes)
	log.Info("Retrieved Name", "Name", retrievedName)

	if retrievedName != transformName {
		t.Fatalf("Retrieved transform name '%s' does not match created name '%s'", retrievedName, transformName)
	}

	// Update the transform
	updateTransform := make(map[string]interface{})
	for k, v := range transform {
		updateTransform[k] = v
	}

	// Change an attribute value to verify the update
	attributes, ok := updateTransform["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("Could not get attributes from transform")
	}
	attributes["sourceName"] = "Updated Workday"

	err = SaveTransform(updateFile, updateTransform)
	if err != nil {
		t.Fatalf("Unable to save update test data: %v", err)
	}

	// Update the transform
	updateCMD := newPutCmd()
	updateBuffer := new(bytes.Buffer)
	updateCMD.SetOut(updateBuffer)
	updateCMD.SetArgs([]string{"/v2024/transforms/" + transformID})
	updateCMD.Flags().Set("body-file", filepath.Join(path, updateFile))
	updateCMD.Flags().Set("jsonpath", "$.attributes.sourceName")

	err = updateCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewUpdateCmd: Unable to execute the command successfully: %v", err)
	}

	// Read the output
	responseBytes, err = io.ReadAll(updateBuffer)
	if err != nil {
		t.Fatalf("Error reading stdout: %v", err)
	}
	putSourceName := string(responseBytes)
	log.Info("PUT Source Name", "Source Name", putSourceName)

	// Verify the update by getting the transform again
	getBuffer.Reset()
	err = getCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewGetCmd: Unable to execute the command successfully after update: %v", err)
	}

	// Read the output
	responseBytes, err = io.ReadAll(getBuffer)
	if err != nil {
		t.Fatalf("Error reading stdout: %v", err)
	}
	retrievedName = string(responseBytes)
	log.Info("Retrieved Name", "Name", retrievedName)

	// Clean up - delete the transform
	deleteCMD := newDeleteCmd()
	deleteBuffer := new(bytes.Buffer)
	deleteCMD.SetOut(deleteBuffer)
	deleteCMD.SetArgs([]string{"/v2024/transforms/" + transformID})

	err = deleteCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewDeleteCmd: Unable to execute the command successfully: %v", err)
	}
}
