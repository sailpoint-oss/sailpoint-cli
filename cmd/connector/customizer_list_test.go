// connector/list_test.go
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

func TestNewCustomizerListCmd_httpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint)), gomock.Nil()).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     http.StatusText(http.StatusBadRequest),
			Body:       io.NopCloser(strings.NewReader("oops")),
		}, nil).
		Times(1)

	cmd := newCustomizerListCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "list customizers failed") {
		t.Fatalf("expected HTTP-error, got %v", err)
	}
}

func TestNewCustomizerListCmd_jsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint)), gomock.Nil()).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(strings.NewReader("not-json")),
		}, nil).
		Times(1)

	cmd := newCustomizerListCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected JSON decode error, got %v", err)
	}
}

func TestNewCustomizerListCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// prepare two customizers
	items := []customizer{
		{ID: "c1", Name: "First"},
		{ID: "c2", Name: "Second"},
	}
	raw, _ := json.Marshal(items)

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint)), gomock.Nil()).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(bytes.NewReader(raw)),
		}, nil).
		Times(1)

	cmd := newCustomizerListCmd(mockClient)
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected Execute error: %v", err)
	}

	out := outBuf.String()
	// check each ID appears
	for _, c := range items {
		if !strings.Contains(out, c.ID) {
			t.Errorf("output missing ID %q, got:\n%s", c.ID, out)
		}
	}
	// check each header uppercased
	for _, col := range customizerColumns {
		uc := strings.ToUpper(col)
		if !strings.Contains(out, uc) {
			t.Errorf("output missing header %q, got:\n%s", uc, out)
		}
	}
}
