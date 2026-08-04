package ui_plugins

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

// enabledInstanceBody is a resolved instance in the ENABLED state, so a disable
// proceeds to the PATCH rather than short-circuiting on idempotency.
const enabledInstanceBody = `{"pluginInstanceId":"pi-1","alias":"my-plugin","name":{"en":"My Plugin"},"created":"2026-05-01T00:00:00Z","state":"ENABLED","slots":[{"slotId":"full-page"}]}`

// disabledPatchResponse is the instance as returned by the PATCH, now DISABLED.
// The `modified` field is unique to the PATCH response so tests can prove that
// --json emits the post-change document rather than the pre-change resolve body.
const disabledPatchResponse = `{"pluginInstanceId":"pi-1","alias":"my-plugin","state":"DISABLED","modified":"2026-07-29T00:00:00Z"}`

func TestRunDisable_ConfirmYesDisables(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: enabledInstanceBody}},
		patchResp: stubResp{status: 200, body: disabledPatchResponse},
	}
	var out bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader("y\n"), &out, &out, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 1 {
		t.Fatalf("expected 1 Patch, got %d", c.patchCalls)
	}
	if !strings.HasSuffix(c.patchURL, "/pi-1") {
		t.Fatalf("expected PATCH on resolved id, got: %s", c.patchURL)
	}
	if got := strings.TrimSpace(string(c.patchBody)); got != `{"state":"DISABLED"}` {
		t.Fatalf("unexpected PATCH body: %s", got)
	}
	if got := c.patchHeaders["Content-Type"]; got != "application/json" {
		t.Fatalf("PATCH must set Content-Type application/json, got: %q", got)
	}
	got := out.String()
	if !strings.Contains(got, "full-page") {
		t.Fatalf("expected confirmation to show slots, got: %s", got)
	}
	if !strings.Contains(got, "Disabled plugin instance pi-1 (alias: my-plugin)") {
		t.Fatalf("expected success message with id and alias, got: %s", got)
	}
}

func TestRunDisable_DeclineDoesNotDisable(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: enabledInstanceBody}},
		patchResp: stubResp{status: 200, body: disabledPatchResponse},
	}
	var out bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader("n\n"), &out, &out, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("declining must not call Patch")
	}
	if !strings.Contains(out.String(), "Cancelled.") {
		t.Fatalf("expected cancellation message, got: %s", out.String())
	}
}

func TestRunDisable_EmptyInputDefaultsToNo(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: enabledInstanceBody}},
		patchResp: stubResp{status: 200, body: disabledPatchResponse},
	}
	var out bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader("\n"), &out, &out, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("empty input must default to No and not disable")
	}
}

func TestRunDisable_ForceSkipsPrompt(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: enabledInstanceBody}},
		patchResp: stubResp{status: 200, body: disabledPatchResponse},
	}
	var out bytes.Buffer

	// Empty stdin: if the prompt were shown, it would default to No and skip the PATCH.
	err := runDisable(context.Background(), c, strings.NewReader(""), &out, &out, "my-plugin", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 1 {
		t.Fatalf("--force should disable without prompting, got %d Patch calls", c.patchCalls)
	}
	if strings.Contains(out.String(), "[y/N]") {
		t.Fatalf("--force should not print a prompt, got: %s", out.String())
	}
}

func TestRunDisable_AlreadyDisabledSkipsPatch(t *testing.T) {
	body := `{"pluginInstanceId":"pi-1","alias":"my-plugin","state":"DISABLED"}`
	c := &seqClient{getQueue: []stubResp{{status: 200, body: body}}}
	var out, errOut bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader(""), &out, &errOut, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("an already-disabled instance must not PATCH")
	}
	if !strings.Contains(out.String(), "already disabled") || !strings.Contains(out.String(), "pi-1") {
		t.Fatalf("expected already-disabled notice on stdout, got: %s", out.String())
	}
	if strings.Contains(out.String(), "[y/N]") {
		t.Fatalf("idempotent path must not prompt, got: %s", out.String())
	}
}

func TestRunDisable_AlreadyDisabledJSON(t *testing.T) {
	body := `{"pluginInstanceId":"pi-1","alias":"my-plugin","state":"DISABLED"}`
	c := &seqClient{getQueue: []stubResp{{status: 200, body: body}}}
	var out, errOut bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader(""), &out, &errOut, "my-plugin", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("an already-disabled instance must not PATCH")
	}
	if !strings.Contains(out.String(), `"pluginInstanceId":"pi-1"`) {
		t.Fatalf("--json stdout should carry the instance JSON, got: %s", out.String())
	}
	if strings.Contains(out.String(), "already disabled") {
		t.Fatalf("--json stdout must be pure JSON, got: %s", out.String())
	}
	if !strings.Contains(errOut.String(), "already disabled") {
		t.Fatalf("expected already-disabled notice on stderr, got: %s", errOut.String())
	}
}

func TestRunDisable_JSONEmitsUpdatedInstance(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: enabledInstanceBody}},
		patchResp: stubResp{status: 200, body: disabledPatchResponse},
	}
	var out, errOut bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader(""), &out, &errOut, "my-plugin", true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// stdout must carry the PATCH response (post-change), not the stale resolve body:
	// `modified` and DISABLED both appear only in the PATCH response.
	if !strings.Contains(out.String(), `"modified"`) || !strings.Contains(out.String(), `"state":"DISABLED"`) {
		t.Fatalf("expected updated instance JSON on stdout, got: %s", out.String())
	}
	if strings.Contains(out.String(), "Disabled plugin instance") {
		t.Fatalf("--json stdout should be pure JSON, got: %s", out.String())
	}
}

func TestRunDisable_NotFoundBeforePrompt(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 404, body: `{"message":"Not Found"}`}}}
	var out bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader("y\n"), &out, &out, "missing", false, false)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("must not PATCH when target not found")
	}
	if strings.Contains(out.String(), "[y/N]") {
		t.Fatalf("must not prompt when target resolution fails, got: %s", out.String())
	}
}

func TestRunDisable_AmbiguousNotDisabledEvenWithForce(t *testing.T) {
	body := `{"conflicts":[{"pluginInstanceId":"pi-1"},{"pluginInstanceId":"pi-2"}]}`
	c := &seqClient{getQueue: []stubResp{{status: 409, body: body}}}
	var out bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader(""), &out, &out, "dup-plugin", true, false)
	if err == nil {
		t.Fatal("expected ambiguity error even with --force")
	}
	if c.patchCalls != 0 {
		t.Fatal("ambiguous alias must never be disabled, even with --force")
	}
	if !strings.Contains(err.Error(), "pi-1") || !strings.Contains(err.Error(), "pi-2") {
		t.Fatalf("expected conflicting IDs in error, got: %v", err)
	}
}

func TestRunDisable_ActiveBundleWarning(t *testing.T) {
	body := `{"pluginInstanceId":"pi-1","alias":"my-plugin","state":"ENABLED","activeAssetBundleId":"ab-9","slots":[{"slotId":"full-page"}]}`
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: body}},
		patchResp: stubResp{status: 200, body: disabledPatchResponse},
	}
	var out bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader("y\n"), &out, &out, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "active asset bundle") {
		t.Fatalf("expected active-deployment warning, got: %s", out.String())
	}
}

func TestRunDisable_UpdateErrorMapping(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: enabledInstanceBody}},
		patchResp: stubResp{status: 403, body: `{"message":"Forbidden"}`},
	}
	var out bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader("y\n"), &out, &out, "my-plugin", false, false)
	if err == nil || !strings.Contains(err.Error(), "not authorized to update") {
		t.Fatalf("expected update forbidden mapping, got: %v", err)
	}
}

func TestRunDisable_TransportError(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: 200, body: enabledInstanceBody}},
		patchResp: stubResp{err: errors.New("connection refused")},
	}
	var out bytes.Buffer

	err := runDisable(context.Background(), c, strings.NewReader("y\n"), &out, &out, "my-plugin", false, false)
	if err == nil || !strings.Contains(err.Error(), "failed to update plugin instance state") {
		t.Fatalf("expected wrapped transport error, got: %v", err)
	}
}
