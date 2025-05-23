// connector/get_test.go
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

func TestNewCustomizerGetCmd_missingRequiredFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Get(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	cmd := newCustomizerGetCmd(mockClient)
	cmd.SetArgs([]string{}) // no -c

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail when -c is missing")
	}
}

func TestNewCustomizerGetCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Minimal JSON for the customizer struct
	cust := map[string]interface{}{"id": "cust-123", "name": "MyCustom"}
	raw, _ := json.Marshal(cust)

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Get(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "cust-123")),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(bytes.NewReader(raw)),
		}, nil).
		Times(1)

	cmd := newCustomizerGetCmd(mockClient)
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetArgs([]string{"-c", "cust-123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected Execute error: %v", err)
	}

	out := outBuf.String()
	// Check that the ID shows up
	if !strings.Contains(out, "cust-123") {
		t.Errorf("output does not contain customizer ID, got:\n%s", out)
	}
	// Check that each column header appears somewhere (tablewriter uppercases them)
	for _, col := range customizerColumns {
		uc := strings.ToUpper(col)
		if !strings.Contains(out, uc) {
			t.Errorf("output does not contain header %q (uppercase %q), got:\n%s", col, uc, out)
		}
	}
}

func TestNewCustomizerGetCmd_httpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Get(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "bad-123")),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     http.StatusText(http.StatusBadRequest),
			Body:       io.NopCloser(bytes.NewReader([]byte("oops"))),
		}, nil).
		Times(1)

	cmd := newCustomizerGetCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "bad-123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error on HTTP status != 200")
	}
	if !strings.Contains(err.Error(), "create customizer failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewCustomizerGetCmd_jsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Get(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "bad-json")),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(bytes.NewReader([]byte("not-json"))),
		}, nil).
		Times(1)

	cmd := newCustomizerGetCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "bad-json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected JSON decode error")
	}
	if !strings.Contains(err.Error(), "invalid character") {
		t.Errorf("unexpected error message: %v", err)
	}
}
