// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package cmd

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint/sp-cli/mocks"
)

func TestNewConnCreateVersionCmd_missingRequiredFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/zip", gomock.Any()).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil).
		Times(0)

	cmd := newConnCreateVersionCmd(client)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail")
	}
}

func TestNewConnCreateVersionCmd_invalidZip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/zip", gomock.Any()).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil).
		Times(0)

	cmd := newConnCreateVersionCmd(client)
	cmd.SetArgs([]string{"-c", "mockConnectorId", "-f", "not-exist.zip"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail")
	}
}
