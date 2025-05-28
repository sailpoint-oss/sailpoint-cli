// connector/customizer_create_version_test.go
package connector

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

// helper: create a minimal in-memory zip file and write it to path
func writeTestZip(path string) error {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	f, err := zw.Create("dummy.txt")
	if err != nil {
		return err
	}
	if _, err := f.Write([]byte("content")); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func TestNewCustomizerCreateVersionCmd_missingFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	cmd := newCustomizerCreateVersionCmd(mockClient)
	// neither -c nor -f
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when flags are missing")
	}

	// just -c
	cmd = newCustomizerCreateVersionCmd(mockClient)
	cmd.SetArgs([]string{"-c", "cust-1"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when -f is missing")
	}

	// just -f
	cmd = newCustomizerCreateVersionCmd(mockClient)
	cmd.SetArgs([]string{"-f", "some.zip"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when -c is missing")
	}
}

func TestNewCustomizerCreateVersionCmd_fileOpenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	cmd := newCustomizerCreateVersionCmd(mockClient)
	cmd.SetArgs([]string{"-c", "cust-1", "-f", "does_not_exist.zip"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected a file-open error, got nil")
	}
	msg := err.Error()
	// accept either Windows or Unix style
	if !strings.Contains(msg, "cannot find the file specified") &&
		!strings.Contains(msg, "no such file or directory") {
		t.Fatalf("unexpected file-open error: %v", err)
	}
}

func TestNewCustomizerCreateVersionCmd_invalidZip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	tmp := t.TempDir()
	path := tmp + "/notazip.zip"
	os.WriteFile(path, []byte("not a zip"), 0o644)

	cmd := newCustomizerCreateVersionCmd(mockClient)
	cmd.SetArgs([]string{"-c", "cust-1", "-f", path})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "zip: not a valid zip") {
		t.Fatalf("expected invalid zip error, got %v", err)
	}
}

func TestNewCustomizerCreateVersionCmd_httpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	tmp := t.TempDir()
	zipPath := tmp + "/test.zip"
	if err := writeTestZip(zipPath); err != nil {
		t.Fatal(err)
	}

	mockClient.
		EXPECT().
		Post(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "cust-1", "versions")),
			gomock.Eq("application/zip"),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Status:     http.StatusText(http.StatusBadRequest),
			Body:       io.NopCloser(strings.NewReader("oops")),
		}, nil).
		Times(1)

	cmd := newCustomizerCreateVersionCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "cust-1", "-f", zipPath})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "upload customizer failed") {
		t.Fatalf("expected HTTP-error, got %v", err)
	}
}

func TestNewCustomizerCreateVersionCmd_jsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	tmp := t.TempDir()
	zipPath := tmp + "/test.zip"
	if err := writeTestZip(zipPath); err != nil {
		t.Fatal(err)
	}

	mockClient.
		EXPECT().
		Post(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "cust-1", "versions")),
			gomock.Eq("application/zip"),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(strings.NewReader("not-json")),
		}, nil).
		Times(1)

	cmd := newCustomizerCreateVersionCmd(mockClient)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "cust-1", "-f", zipPath})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected JSON decode error, got %v", err)
	}
}

func TestNewCustomizerCreateVersionCmd_success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	tmp := t.TempDir()
	zipPath := tmp + "/test.zip"
	if err := writeTestZip(zipPath); err != nil {
		t.Fatal(err)
	}

	// prepare version response with integer Version
	cv := customizerVersion{Version: 2}
	rawOut, _ := json.Marshal(cv)

	mockClient.
		EXPECT().
		Post(
			gomock.Any(),
			gomock.Eq(util.ResourceUrl(connectorCustomizersEndpoint, "cust-1", "versions")),
			gomock.Eq("application/zip"),
			gomock.Any(),
			gomock.Nil(),
		).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       io.NopCloser(bytes.NewReader(rawOut)),
		}, nil).
		Times(1)

	cmd := newCustomizerCreateVersionCmd(mockClient)
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetArgs([]string{"-c", "cust-1", "-f", zipPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected Execute error: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, "2") {
		t.Errorf("output missing version, got:\n%s", out)
	}
	// uppercase headers
	for _, col := range customizerVersionColumns {
		if !strings.Contains(out, strings.ToUpper(col)) {
			t.Errorf("output missing header %s, got:\n%s", col, out)
		}
	}
}
