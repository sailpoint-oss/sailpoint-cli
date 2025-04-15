// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
)

func TestAccountCreateWithIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:account:create","config":{},` +
		`"input":{"identity":"john.doe","attributes":{}}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i)), nil).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeAccountCreateCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe", "-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}

func TestAccountCreateWithIdentityAndAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:account:create","config":{},` +
		`"input":{"identity":"john.doe","attributes":{"foo":"bar"}}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i)), nil).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeAccountCreateCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe", "-c", "test-connector", "--config-json", "{}", "--attributes", `{"foo":"bar"}`})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}
