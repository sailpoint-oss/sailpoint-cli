// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
)

func TestAccounUpdateWithoutInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)

	cmd := newConnInvokeAccountUpdateCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err == nil {
		t.Errorf("failed to detect error: updating account without identity")
	}
}

func TestAccountUpdateWithIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:account:update","config":{},"input":{"identity":"john.doe","key":{"simple":{"id":"john.doe"}},"changes":[]}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i))).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeAccountUpdateCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe", "-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}

func TestAccountUpdateWithIdentityAndChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := `[{"op":"Add","attribute":"location","value":"austin"}]`
	i := fmt.Sprintf(`{"connectorRef":"test-connector","tag":"latest","type":"std:account:update","config":{},"input":{"identity":"john.doe","key":{"simple":{"id":"john.doe"}},"changes":%s}}`, c)

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i))).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeAccountUpdateCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe", "-c", "test-connector", "--config-json", "{}", "--changes", c})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}
