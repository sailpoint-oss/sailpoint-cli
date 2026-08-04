package ui_plugins

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

// disabledInstanceBody is a resolved instance in the DISABLED state, so an enable
// proceeds to the PATCH rather than short-circuiting on idempotency.
const disabledInstanceBody = `{"pluginInstanceId":"pi-1","alias":"my-plugin","name":{"en":"My Plugin"},"created":"2026-05-01T00:00:00Z","state":"DISABLED"}`

// enabledPatchResponse is the instance as returned by the PATCH, now ENABLED. The
// `modified` field is unique to the PATCH response so tests can prove that --json
// emits the post-change document rather than the pre-change resolve body.
const enabledPatchResponse = `{"pluginInstanceId":"pi-1","alias":"my-plugin","state":"ENABLED","modified":"2026-07-29T00:00:00Z"}`

func TestRunEnable_Enables(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: disabledInstanceBody}},
		patchResp: stubResp{status: 200, body: enabledPatchResponse},
	}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "my-plugin", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 1 {
		t.Fatalf("expected 1 Patch, got %d", c.patchCalls)
	}
	if !strings.HasSuffix(c.patchURL, "/pi-1") {
		t.Fatalf("expected PATCH on resolved id, got: %s", c.patchURL)
	}
	if got := strings.TrimSpace(string(c.patchBody)); got != `{"state":"ENABLED"}` {
		t.Fatalf("unexpected PATCH body: %s", got)
	}
	if got := c.patchHeaders["Content-Type"]; got != "application/json" {
		t.Fatalf("PATCH must set Content-Type application/json, got: %q", got)
	}
	if !strings.Contains(out.String(), "Enabled plugin instance pi-1 (alias: my-plugin)") {
		t.Fatalf("expected success message with id and alias, got: %s", out.String())
	}
}

func TestRunEnable_NoConfirmationPrompt(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: disabledInstanceBody}},
		patchResp: stubResp{status: 200, body: enabledPatchResponse},
	}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "my-plugin", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// enable is non-destructive and intentionally has no confirmation.
	if strings.Contains(out.String(), "[y/N]") || strings.Contains(out.String(), "You are about to") {
		t.Fatalf("enable must not prompt for confirmation, got: %s", out.String())
	}
}

func TestRunEnable_AlreadyEnabledSkipsPatch(t *testing.T) {
	body := `{"pluginInstanceId":"pi-1","alias":"my-plugin","state":"ENABLED"}`
	c := &seqClient{getQueue: []stubResp{{status: 200, body: body}}}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "my-plugin", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("an already-enabled instance must not PATCH")
	}
	if !strings.Contains(out.String(), "already enabled") || !strings.Contains(out.String(), "pi-1") {
		t.Fatalf("expected already-enabled notice on stdout, got: %s", out.String())
	}
}

func TestRunEnable_AlreadyEnabledJSON(t *testing.T) {
	body := `{"pluginInstanceId":"pi-1","alias":"my-plugin","state":"ENABLED"}`
	c := &seqClient{getQueue: []stubResp{{status: 200, body: body}}}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "my-plugin", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("an already-enabled instance must not PATCH")
	}
	if !strings.Contains(out.String(), `"pluginInstanceId":"pi-1"`) {
		t.Fatalf("--json stdout should carry the instance JSON, got: %s", out.String())
	}
	if strings.Contains(out.String(), "already enabled") {
		t.Fatalf("--json stdout must be pure JSON, got: %s", out.String())
	}
	if !strings.Contains(errOut.String(), "already enabled") {
		t.Fatalf("expected already-enabled notice on stderr, got: %s", errOut.String())
	}
}

func TestRunEnable_JSONEmitsUpdatedInstance(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: disabledInstanceBody}},
		patchResp: stubResp{status: 200, body: enabledPatchResponse},
	}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "my-plugin", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// stdout must carry the PATCH response (post-change), not the stale resolve body.
	if !strings.Contains(out.String(), `"modified"`) || !strings.Contains(out.String(), `"state":"ENABLED"`) {
		t.Fatalf("expected updated instance JSON on stdout, got: %s", out.String())
	}
	if strings.Contains(out.String(), "Enabled plugin instance") {
		t.Fatalf("--json stdout should be pure JSON, got: %s", out.String())
	}
}

func TestRunEnable_NotFound(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 404, body: `{"message":"Not Found"}`}}}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "missing", false)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("must not PATCH when target not found")
	}
}

func TestRunEnable_AmbiguousAlias(t *testing.T) {
	body := `{"conflicts":[{"pluginInstanceId":"pi-1"},{"pluginInstanceId":"pi-2"}]}`
	c := &seqClient{getQueue: []stubResp{{status: 409, body: body}}}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "dup-plugin", false)
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	if c.patchCalls != 0 {
		t.Fatal("ambiguous alias must never be enabled")
	}
	if !strings.Contains(err.Error(), "pi-1") || !strings.Contains(err.Error(), "pi-2") {
		t.Fatalf("expected conflicting IDs in error, got: %v", err)
	}
}

func TestRunEnable_UpdateErrorMapping(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: disabledInstanceBody}},
		patchResp: stubResp{status: 403, body: `{"message":"Forbidden"}`},
	}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "my-plugin", false)
	if err == nil || !strings.Contains(err.Error(), "not authorized to update") {
		t.Fatalf("expected update forbidden mapping, got: %v", err)
	}
}

func TestRunEnable_TransportError(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: disabledInstanceBody}},
		patchResp: stubResp{err: errors.New("connection refused")},
	}
	var out, errOut bytes.Buffer

	err := runEnable(context.Background(), c, &out, &errOut, "my-plugin", false)
	if err == nil || !strings.Contains(err.Error(), "failed to update plugin instance state") {
		t.Fatalf("expected wrapped transport error, got: %v", err)
	}
}
