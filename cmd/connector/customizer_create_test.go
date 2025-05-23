// connector/create_test.go
package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

func TestNewCustomizerCreateCmd_missingArg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	// Post should never be called when arg is missing
	mockClient.
		EXPECT().
		Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	cmd := newCustomizerCreateCmd(mockClient)
	cmd.SetArgs([]string{}) // no <customizer-name>

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when customizer name arg is missing")
	}
}

func TestNewCustomizerCreateCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// prepare JSON input and response
	input := customizer{Name: "MyCustom"}
	rawIn, _ := json.Marshal(input)
	created := customizer{ID: "c-123", Name: "MyCustom"}
	rawOut, _ := json.Marshal(created)

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Post(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint)),
			gomock.Eq("application/json"),
			gomock.Any(), // reader matching rawIn
			gomock.Nil(),
		).
		DoAndReturn(func(ctx context.Context, url, cType string, body io.Reader, headers map[string]string) (*http.Response, error) {
			// verify body content
			got, _ := io.ReadAll(body)
			if !bytes.Equal(got, rawIn) {
				return nil, fmt.Errorf("unexpected request body: %s", got)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Body:       io.NopCloser(bytes.NewReader(rawOut)),
			}, nil
		}).
		Times(1)

	cmd := newCustomizerCreateCmd(mockClient)
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetArgs([]string{"MyCustom"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected Execute error: %v", err)
	}

	out := outBuf.String()
	// should contain returned ID
	if !strings.Contains(out, "c-123") {
		t.Errorf("output does not contain ID, got:\n%s", out)
	}
	// headers are uppercased by tablewriter
	for _, col := range customizerColumns {
		if !strings.Contains(out, strings.ToUpper(col)) {
			t.Errorf("output missing header %s, got:\n%s", col, out)
		}
	}
}

func TestNewCustomizerCreateCmd_httpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Post(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint)),
			gomock.Any(), gomock.Any(), gomock.Any(),
		).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     http.StatusText(http.StatusBadRequest),
			Body:       io.NopCloser(strings.NewReader("oops")),
		}, nil).
		Times(1)

	cmd := newCustomizerCreateCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"MyCustom"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error on HTTP status != 200")
	}
	if !strings.Contains(err.Error(), "create customizer failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewCustomizerCreateCmd_jsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Post(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint)),
			gomock.Any(), gomock.Any(), gomock.Any(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(strings.NewReader("not-json")),
		}, nil).
		Times(1)

	cmd := newCustomizerCreateCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"MyCustom"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected JSON decode error")
	}
	if !strings.Contains(err.Error(), "invalid character") {
		t.Errorf("unexpected error: %v", err)
	}
}
