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

func TestSourceDataReadWithInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:source-data:read","config":{},"input":{"sourceDataKey":"john.doe","queryInput":{}}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i)), nil).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeSourceDataReadCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe", "-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err == nil {
		t.Errorf("failed to detect error: reading source data discover")
	}
}
