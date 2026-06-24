package clierror

import (
	"errors"
	"strings"
	"testing"
)

func TestExitCodeUsesTypedErrorCode(t *testing.T) {
	if got := ExitCode(Usage("bad flag")); got != ExitUsage {
		t.Fatalf("expected usage exit code, got %d", got)
	}
	if got := ExitCode(errors.New("plain")); got != ExitGeneral {
		t.Fatalf("expected general exit code, got %d", got)
	}
	if got := ExitCode(nil); got != 0 {
		t.Fatalf("expected success exit code, got %d", got)
	}
}

func TestAPIStatusRedactsBody(t *testing.T) {
	err := APIStatus(401, "401 Unauthorized", []byte(`{"access_token":"secret","message":"denied"}`))
	if ExitCode(err) != ExitAPI {
		t.Fatalf("expected API exit code")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("expected secret to be redacted from %q", err.Error())
	}
	if !strings.Contains(err.Error(), "denied") {
		t.Fatalf("expected non-sensitive API detail to remain visible")
	}
}
