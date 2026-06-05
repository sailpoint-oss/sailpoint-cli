package spconfig

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/testutil"
)

var spConfigJobIDPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

func TestSPConfigExportStatusDownloadLifecycle(t *testing.T) {
	testutil.RequireLiveCredentials(t)

	description := testutil.UniqueName("spconfig")
	exportCmd := newExportCommand()
	exportCmd.Flags().Set("description", description)
	exportCmd.Flags().Set("include", "TRANSFORM")

	exportOut, err := captureStdout(func() error {
		return exportCmd.Execute()
	})
	if err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("spconfig export failed: %v", err)
	}

	jobID := spConfigJobIDPattern.FindString(exportOut)
	if jobID == "" {
		t.Fatalf("expected SPConfig export job ID in output %q", exportOut)
	}

	statusCmd := newStatusCommand()
	statusCmd.Flags().Set("export", jobID)
	if _, err := captureStdout(func() error {
		return statusCmd.Execute()
	}); err != nil {
		t.Fatalf("spconfig status failed: %v", err)
	}

	dir := t.TempDir()
	downloadCmd := newDownloadCommand()
	downloadCmd.Flags().Set("export", jobID)
	downloadCmd.Flags().Set("folder-path", dir)
	if err := downloadCmd.Execute(); err != nil {
		t.Fatalf("spconfig download failed: %v", err)
	}

	expectedPath := filepath.Join(dir, "spconfig-export-"+jobID+".json")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected downloaded export at %s: %v", expectedPath, err)
	}
}

func captureStdout(fn func() error) (string, error) {
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	runErr := fn()
	_ = writer.Close()

	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, reader)
	_ = reader.Close()
	if runErr != nil {
		return buf.String(), runErr
	}
	return buf.String(), copyErr
}
