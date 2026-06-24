package api

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func resetAPIOutputTestState(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func TestEnsureSuccessRejectsNon2xx(t *testing.T) {
	err := ensureSuccess(&http.Response{StatusCode: http.StatusNotFound, Status: "404 Not Found"}, []byte(`{"detail":"missing"}`))
	if err == nil {
		t.Fatal("expected non-2xx response to fail")
	}
	if !strings.Contains(err.Error(), "404 Not Found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteResponseWritesStatusToStderrForHumanOutput(t *testing.T) {
	resetAPIOutputTestState(t)

	cmd := &cobra.Command{}
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	if err := writeResponse(cmd, []byte(`{"id":"123"}`), "200 OK", ""); err != nil {
		t.Fatalf("writeResponse returned error: %v", err)
	}

	if !strings.Contains(out.String(), `{"id":"123"}`) {
		t.Fatalf("expected body on stdout, got %q", out.String())
	}
	if !strings.Contains(errOut.String(), "Status: 200 OK") {
		t.Fatalf("expected status on stderr, got %q", errOut.String())
	}
}

func TestWriteResponseOmitsStatusForMachineReadableOutput(t *testing.T) {
	resetAPIOutputTestState(t)
	viper.Set("output", "json")

	cmd := &cobra.Command{}
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	if err := writeResponse(cmd, []byte(`{"id":"123"}`), "200 OK", ""); err != nil {
		t.Fatalf("writeResponse returned error: %v", err)
	}

	if !strings.Contains(out.String(), `"id": "123"`) {
		t.Fatalf("expected structured JSON on stdout, got %q", out.String())
	}
	if errOut.String() != "" {
		t.Fatalf("expected no status on stderr for machine-readable output, got %q", errOut.String())
	}
}
