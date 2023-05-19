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

func TestEntitlementListWithoutInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)

	cmd := newConnInvokeEntitlementListCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err == nil {
		t.Errorf("failed to detect error: listing entitlements without type")
	}
}

func TestEntitlementListWithType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:entitlement:list","config":{},"input":{"type":"group"}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i))).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeEntitlementListCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"-c", "test-connector", "--config-json", "{}", "--type", "group"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}
