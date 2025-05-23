// cmd/connector/link_test.go
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

func TestNewCustomizerLinkCmd_missingFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	// Patch should never be called if flags missing
	mockClient.EXPECT().
		Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	cmd := newCustomizerLinkCmd(mockClient)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when -c and/or -i missing")
	}
}

func TestNewCustomizerLinkCmd_httpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Patch(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-1", "link")),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     http.StatusText(http.StatusBadRequest),
			Body:       io.NopCloser(strings.NewReader("oops")),
		}, nil).
		Times(1)

	cmd := newCustomizerLinkCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "cust-1", "-i", "inst-1"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "link customizer failed") {
		t.Fatalf("expected HTTP error, got %v", err)
	}
}

func TestNewCustomizerLinkCmd_jsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Patch(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-2", "link")),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(strings.NewReader("not-json")),
		}, nil).
		Times(1)

	cmd := newCustomizerLinkCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "cust-1", "-i", "inst-2"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected JSON decode error, got %v", err)
	}
}

func TestNewCustomizerLinkCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// simulate a successful response with a valid instance
	inst := instance{ID: "inst-3", Name: "LinkedInst"}
	raw, _ := json.Marshal(inst)

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Patch(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorInstancesEndpoint, "inst-3", "link")),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(bytes.NewReader(raw)),
		}, nil).
		Times(1)

	cmd := newCustomizerLinkCmd(mockClient)
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetArgs([]string{"-c", "cust-123", "-i", "inst-3"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, inst.ID) {
		t.Errorf("output missing instance ID, got:\n%s", out)
	}
	for _, col := range instanceColumns {
		if !strings.Contains(out, strings.ToUpper(col)) {
			t.Errorf("output missing header %q, got:\n%s", col, out)
		}
	}
}
