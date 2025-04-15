// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
	"github.com/spf13/cobra"
)

func TestAccountListWithoutInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:account:list","config":{},"input":{}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i)), nil).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeAccountListCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"-c", "test-connector", "--config-json", "{}"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}

func TestAccountListWithState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	i := `{"connectorRef":"test-connector","tag":"latest","type":"std:account:list","config":{},"input":{"stateful":true,"stateId":"123"}}`

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i)), nil).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeAccountListCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"-c", "test-connector", "--config-json", "{}", "--stateful", "--stateId", "123"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}

func TestAccountListWithSchema(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	schema := `{"attributes":[{"description":"The identity of the account","name":"identity","type":"string"}]}`
	i := fmt.Sprintf(`{"connectorRef":"test-connector","tag":"latest","type":"std:account:list","config":{},"input":{"schema":%s}}`, schema)

	file, err := ioutil.TempFile("", "config.json")
	if err != nil {
		t.Errorf("failed to create tempfile %s", err)
	}
	defer os.Remove(file.Name())

	file.Write([]byte(schema))
	defer file.Close()

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Post(gomock.Any(), gomock.Any(), "application/json", bytes.NewReader([]byte(i)), nil).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil)

	cmd := newConnInvokeAccountListCmd(client)
	addRequiredFlagsFromParentCmd(cmd)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"-c", "test-connector", "--config-json", "{}", "--schema", file.Name()})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("command failed with err: %s", err)
	}
}

func addRequiredFlagsFromParentCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("id", "c", "", "")
	cmd.PersistentFlags().StringP("version", "v", "", "")
	cmd.PersistentFlags().StringP("conn-endpoint", "e", connectorsEndpoint, "")
	cmd.PersistentFlags().StringP("config-json", "", "", "Config JSON to use for commands")
}
