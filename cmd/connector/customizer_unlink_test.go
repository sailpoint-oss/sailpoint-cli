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

	// Patch must not be called
	mockClient.
		EXPECT().
		Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	cmd := newCustomizerUnlinkCmd(mockClient)
	cmd.SetArgs([]string{}) // no -i

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when required flag is missing")
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
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-1", "unlink")),
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
		t.Fatalf("expected HTTP error, got %v", err)
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
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-2", "unlink")),
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
		t.Fatalf("expected JSON decode error, got %v", err)
	}
}

func TestNewCustomizerUnlinkCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	// prepare a fake instance response
	inst := instance{ID: "inst-3", Name: "FooInst"}
	raw, _ := json.Marshal(inst)

	mockClient.
		EXPECT().
		Patch(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-3", "unlink")),
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
		t.Fatalf("unexpected Execute error: %v", err)
	}

	out := outBuf.String()
	// ID must appear
	if !strings.Contains(out, inst.ID) {
		t.Errorf("output missing instance ID, got:\n%s", out)
	}
	// uppercase headers
	for _, col := range instanceColumns {
		uc := strings.ToUpper(col)
		if !strings.Contains(out, uc) {
			t.Errorf("output missing header %q, got:\n%s", uc, out)
		}
	}
}
