package cmd

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint/sp-cli/mocks"
)

func TestNewConnGetCmd_missingRequiredFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil).
		Times(0)

	cmd := newConnGetCmd(client)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail")
	}
}

func TestNewConnGetCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	client.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("{}")))}, nil).
		Times(1)

	cmd := newConnGetCmd(client)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"-c", "test-connector"})
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
