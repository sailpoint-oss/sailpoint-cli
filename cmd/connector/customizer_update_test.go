// connector/update_test.go
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

func TestNewCustomizerUpdateCmd_missingFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	// neither -c nor -n
	cmd := newCustomizerUpdateCmd(mockClient)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when flags are missing")
	}

	// just -c
	cmd = newCustomizerUpdateCmd(mockClient)
	cmd.SetArgs([]string{"-c", "cust-1"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when -n is missing")
	}

	// just -n
	cmd = newCustomizerUpdateCmd(mockClient)
	cmd.SetArgs([]string{"-n", "NewName"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when -c is missing")
	}
}

func TestNewCustomizerUpdateCmd_httpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	// stub 400 response
	mockClient.
		EXPECT().
		Put(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "cust-1")),
			gomock.Eq("application/json"),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     http.StatusText(http.StatusBadRequest),
			Body:       io.NopCloser(strings.NewReader("oops")),
		}, nil).
		Times(1)

	cmd := newCustomizerUpdateCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "cust-1", "-n", "NewName"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "create customizer failed") {
		t.Fatalf("expected HTTP-error, got %v", err)
	}
}

func TestNewCustomizerUpdateCmd_jsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	// stub 200 but invalid JSON
	mockClient.
		EXPECT().
		Put(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "cust-1")),
			gomock.Eq("application/json"),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(strings.NewReader("not-json")),
		}, nil).
		Times(1)

	cmd := newCustomizerUpdateCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "cust-1", "-n", "NewName"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected JSON decode error, got %v", err)
	}
}

func TestNewCustomizerUpdateCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	// prepare input & output
	updated := customizer{ID: "cust-1", Name: "NewName"}
	rawOut, _ := json.Marshal(updated)

	mockClient.
		EXPECT().
		Put(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "cust-1")),
			gomock.Eq("application/json"),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(bytes.NewReader(rawOut)),
		}, nil).
		Times(1)

	cmd := newCustomizerUpdateCmd(mockClient)
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetArgs([]string{"-c", "cust-1", "-n", "NewName"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected Execute error: %v", err)
	}

	out := outBuf.String()
	// check updated ID and Name appear
	if !strings.Contains(out, updated.ID) || !strings.Contains(out, updated.Name) {
		t.Errorf("output missing updated data, got:\n%s", out)
	}
	// headers uppercased
	for _, col := range customizerColumns {
		if !strings.Contains(out, strings.ToUpper(col)) {
			t.Errorf("output missing header %q, got:\n%s", col, out)
		}
	}
}
