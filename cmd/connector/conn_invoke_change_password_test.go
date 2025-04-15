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

func TestChangePasswordWithoutInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	term := mocks.NewMockTerminal(ctrl)

	cmd := newConnInvokeChangePasswordCmd(client, term)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err == nil {
		t.Errorf("failed to detect error: changing password without identity")
	}
}

func TestChangePasswordWithIdentityAndPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:change-password","config":{},` +
		`"input":{"identity":"john.doe","key":{"simple":{"id":"john.doe"}},"password":"password"}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i)), nil).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	term := mocks.NewMockTerminal(ctrl)
	term.EXPECT().
		PromptPassword(gomock.Any()).
		Return("password", nil)

	cmd := newConnInvokeChangePasswordCmd(client, term)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe", "-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}

func TestChangePasswordWithIdentityAndPasswordAndUniqueId(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:change-password","config":{},` +
		`"input":{"identity":"john.doe","key":{"compound":{"lookupId":"john.doe","uniqueId":"12345"}},"password":"password"}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i)), nil).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	term := mocks.NewMockTerminal(ctrl)
	term.EXPECT().
		PromptPassword(gomock.Any()).
		Return("password", nil)

	cmd := newConnInvokeChangePasswordCmd(client, term)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe", "12345", "-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}
