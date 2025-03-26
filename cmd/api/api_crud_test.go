// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"bytes"
	"encoding/json"
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

	err = createCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewCreateCmd: Unable to execute the command successfully: %v", err)
	}

	// Capture the response bytes before logging
	responseBytes := createBuffer.Bytes()

	// Extract just the JSON part of the response (before the Status line)
	lines := bytes.Split(responseBytes, []byte("\n"))
	var jsonBytes []byte
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("Status:")) {
			break
		}
		if len(line) > 0 {
			jsonBytes = append(jsonBytes, line...)
			jsonBytes = append(jsonBytes, '\n')
		}
	}

	// Parse the response to get the transform ID
	var response map[string]interface{}
	err = json.Unmarshal(jsonBytes, &response)
	if err != nil {
		t.Fatalf("Error parsing response: %v", err)
	}

	transformID, ok := response["id"].(string)
	if !ok {
		t.Fatal("Could not get transform ID from response")
	}

	// Validate the transform was created by getting it
	getCMD := newGetCmd()
	getBuffer := new(bytes.Buffer)
	getCMD.SetOut(getBuffer)
	getCMD.SetArgs([]string{"/v2024/transforms/" + transformID})

	err = getCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewGetCmd: Unable to execute the command successfully: %v", err)
	}

	// Capture the response bytes before logging
	getResponseBytes := getBuffer.Bytes()

	// Extract just the JSON part of the response
	lines = bytes.Split(getResponseBytes, []byte("\n"))
	jsonBytes = nil
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("Status:")) {
			break
		}
		if len(line) > 0 {
			jsonBytes = append(jsonBytes, line...)
			jsonBytes = append(jsonBytes, '\n')
		}
	}

	// Verify the retrieved transform matches what we created
	var getResponse map[string]interface{}
	err = json.Unmarshal(jsonBytes, &getResponse)
	if err != nil {
		t.Fatalf("Error parsing get response: %v", err)
	}

	// Verify the name matches
	retrievedName, ok := getResponse["name"].(string)
	if !ok || retrievedName != transformName {
		t.Fatalf("Retrieved transform name '%s' does not match created name '%s'", retrievedName, transformName)
	}

	// Update the transform
	updateTransform := make(map[string]interface{})
	for k, v := range transform {
		updateTransform[k] = v
	}

	// Change a value to verify the update
	updateTransform["name"] = "Updated " + transformName

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

	err = updateCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewUpdateCmd: Unable to execute the command successfully: %v", err)
	}

	// Verify the update by getting the transform again
	getBuffer.Reset()
	err = getCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewGetCmd: Unable to execute the command successfully after update: %v", err)
	}

	// Capture the response bytes before logging
	getResponseBytes = getBuffer.Bytes()

	// Extract just the JSON part of the response
	lines = bytes.Split(getResponseBytes, []byte("\n"))
	jsonBytes = nil
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("Status:")) {
			break
		}
		if len(line) > 0 {
			jsonBytes = append(jsonBytes, line...)
			jsonBytes = append(jsonBytes, '\n')
		}
	}

	// Verify the retrieved transform matches our updates
	err = json.Unmarshal(jsonBytes, &getResponse)
	if err != nil {
		t.Fatalf("Error parsing get response after update: %v", err)
	}

	// Verify the name was updated
	retrievedName, ok = getResponse["name"].(string)
	if !ok || retrievedName != "Updated "+transformName {
		t.Fatalf("Retrieved transform name '%s' does not match updated name 'Updated %s'", retrievedName, transformName)
	}

	// Clean up - delete the transform
	deleteCMD := newDeleteCmd()
	deleteCMD.SetArgs([]string{"/v2024/transforms/" + transformID})

	err = deleteCMD.Execute()
	if err != nil {
		t.Fatalf("TestNewDeleteCmd: Unable to execute the command successfully: %v", err)
	}
}
