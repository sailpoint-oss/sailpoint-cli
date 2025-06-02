// connector/unlink_test.go
package connector

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

func TestNewCustomizerUnlinkCmd_missingFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	// Patch should not be called when the required flag is missing
	mockClient.
		EXPECT().
		Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	cmd := newCustomizerUnlinkCmd(mockClient)
	cmd.SetArgs([]string{}) // no -i

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when -i is missing")
	}
}

func TestNewCustomizerUnlinkCmd_httpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Patch(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-1")),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     http.StatusText(http.StatusBadRequest),
			Body:       io.NopCloser(strings.NewReader("oops")),
		}, nil).
		Times(1)

	cmd := newCustomizerUnlinkCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-i", "inst-1"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unlink customizer failed") {
		t.Fatalf("expected HTTP‐error, got %v", err)
	}
}

func TestNewCustomizerUnlinkCmd_jsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Patch(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-2")),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(strings.NewReader("not-json")),
		}, nil).
		Times(1)

	cmd := newCustomizerUnlinkCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-i", "inst-2"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected JSON‐decode error, got %v", err)
	}
}

func TestNewCustomizerUnlinkCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// prepare a fake instance response
	inst := instance{ID: "inst-3", Name: "UnlinkedInst"}
	raw, _ := json.Marshal(inst)

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Patch(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-3")),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(bytes.NewReader(raw)),
		}, nil).
		Times(1)

	cmd := newCustomizerUnlinkCmd(mockClient)
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetArgs([]string{"-i", "inst-3"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := outBuf.String()
	// The table should contain the ID
	if !strings.Contains(out, inst.ID) {
		t.Errorf("output missing ID %q, got:\n%s", inst.ID, out)
	}
	// And uppercase headers
	for _, col := range instanceColumns {
		if !strings.Contains(out, strings.ToUpper(col)) {
			t.Errorf("output missing header %q, got:\n%s", strings.ToUpper(col), out)
		}
	}
}
