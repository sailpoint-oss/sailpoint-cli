// connector/delete_test.go
package connector

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

func TestNewCustomizerDeleteCmd_missingRequiredFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	// Delete should never be called when -c is missing
	mockClient.
		EXPECT().
		Delete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	cmd := newCustomizerDeleteCmd(mockClient)
	cmd.SetArgs([]string{}) // no -c

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail when -c is missing")
	}
}

func TestNewCustomizerDeleteCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)
	mockClient.
		EXPECT().
		Delete(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "cust-123")),
			gomock.Nil(), gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusNoContent,
			Status:     http.StatusText(http.StatusNoContent),
			Body:       io.NopCloser(bytes.NewReader(nil)),
		}, nil).
		Times(1)

	cmd := newCustomizerDeleteCmd(mockClient)
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetArgs([]string{"-c", "cust-123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected Execute error: %v", err)
	}

	out := outBuf.String()
	expected := "connector customizer cust-123 deleted."
	if !strings.Contains(out, expected) {
		t.Errorf("output %q does not contain %q", out, expected)
	}
}
