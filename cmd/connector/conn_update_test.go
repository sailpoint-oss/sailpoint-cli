// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package connector

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
)

func TestNewConnUpdateCmd_missingRequiredFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)

	cmd := newConnUpdateCmd(client)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{})
	cmd.PersistentFlags().StringP("conn-endpoint", "e", connectorsEndpoint, "Override connectors endpoint")

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail")
	}
}

func TestNewConnUpdateCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Put(gomock.Any(), gomock.Any(), "application/json", gomock.Any(), nil).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"id": "123"}`))),
		}, nil)

	cmd := newConnUpdateCmd(client)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--id", "mockConnectorId", "--alias", "newConnectorAlias"})
	cmd.PersistentFlags().StringP("conn-endpoint", "e", connectorsEndpoint, "Override connectors endpoint")

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("error execute cmd: %v", err)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("error read out: %v", err)
	}

	if len(string(out)) == 0 {
		t.Errorf("error empty out")
	}
}
