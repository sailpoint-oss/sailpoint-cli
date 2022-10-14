// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package transform

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/mocks"
)

func TestNewDeleteCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Delete(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(bytes.NewReader([]byte("")))}, nil).
		Times(1)

	cmd := newDeleteCmd(client)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"03d5187b-ab96-402c-b5a1-40b74285d77b"})
	cmd.PersistentFlags().StringP("transforms-endpoint", "e", transformsEndpoint, "Override transforms endpoint")

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("TestNewCreateCmd: Unable to execute the command successfully: %v", err)
	}
}
