package ui_plugins

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

const resolvedAliasBody = `{"pluginInstanceId":"pi-1","alias":"my-plugin","name":{"en":"My Plugin"},"created":"2026-05-01T00:00:00Z","slots":[{"slotId":"full-page"}]}`

func TestRunDelete_ConfirmYesDeletes(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: resolvedAliasBody}},
		deleteResp: stubResp{status: 204},
	}
	var out bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader("y\n"), &out, &out, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.deleteCalls != 1 {
		t.Fatalf("expected 1 Delete, got %d", c.deleteCalls)
	}
	if !strings.HasSuffix(c.deleteURL, "/pi-1") {
		t.Fatalf("expected DELETE on resolved id, got: %s", c.deleteURL)
	}
	got := out.String()
	if !strings.Contains(got, "full-page") {
		t.Fatalf("expected confirmation to show slots, got: %s", got)
	}
	if !strings.Contains(got, "Deleted plugin instance pi-1") || !strings.Contains(got, "my-plugin") {
		t.Fatalf("expected success message with id and alias, got: %s", got)
	}
}

func TestRunDelete_DeclineDoesNotDelete(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: resolvedAliasBody}},
		deleteResp: stubResp{status: 204},
	}
	var out bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader("n\n"), &out, &out, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.deleteCalls != 0 {
		t.Fatal("declining must not call Delete")
	}
	if !strings.Contains(out.String(), "Cancelled.") {
		t.Fatalf("expected cancellation message, got: %s", out.String())
	}
}

func TestRunDelete_EmptyInputDefaultsToNo(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: resolvedAliasBody}},
		deleteResp: stubResp{status: 204},
	}
	var out bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader("\n"), &out, &out, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.deleteCalls != 0 {
		t.Fatal("empty input must default to No and not delete")
	}
}

func TestRunDelete_ForceSkipsPrompt(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: resolvedAliasBody}},
		deleteResp: stubResp{status: 204},
	}
	var out bytes.Buffer

	// Empty stdin: if the prompt were shown, it would default to No and skip Delete.
	err := runDelete(context.Background(), c, strings.NewReader(""), &out, &out, "my-plugin", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.deleteCalls != 1 {
		t.Fatalf("--force should delete without prompting, got %d Delete calls", c.deleteCalls)
	}
	if strings.Contains(out.String(), "[y/N]") {
		t.Fatalf("--force should not print a prompt, got: %s", out.String())
	}
}

func TestRunDelete_AmbiguousNotDeletedEvenWithForce(t *testing.T) {
	body := `{"conflicts":[{"pluginInstanceId":"pi-1"},{"pluginInstanceId":"pi-2"}]}`
	c := &seqClient{getQueue: []stubResp{{status: 409, body: body}}}
	var out bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader(""), &out, &out, "dup-plugin", true, false)
	if err == nil {
		t.Fatal("expected ambiguity error even with --force")
	}
	if c.deleteCalls != 0 {
		t.Fatal("ambiguous alias must never be deleted, even with --force")
	}
	if !strings.Contains(err.Error(), "pi-1") || !strings.Contains(err.Error(), "pi-2") {
		t.Fatalf("expected conflicting IDs in error, got: %v", err)
	}
}

func TestRunDelete_NotFoundBeforePrompt(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 404, body: `{"message":"Not Found"}`}}}
	var out bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader("y\n"), &out, &out, "missing", false, false)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got: %v", err)
	}
	if c.deleteCalls != 0 {
		t.Fatal("must not delete when target not found")
	}
	if strings.Contains(out.String(), "[y/N]") {
		t.Fatalf("must not prompt when target resolution fails, got: %s", out.String())
	}
}

func TestRunDelete_ActiveBundleWarning(t *testing.T) {
	body := `{"pluginInstanceId":"pi-1","alias":"my-plugin","activeAssetBundleId":"ab-9","slots":[{"slotId":"full-page"}]}`
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: body}},
		deleteResp: stubResp{status: 204},
	}
	var out bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader("y\n"), &out, &out, "my-plugin", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "active asset bundle") {
		t.Fatalf("expected active-deployment warning, got: %s", out.String())
	}
}

func TestRunDelete_JSONEmitsInstanceOnStdout(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: resolvedAliasBody}},
		deleteResp: stubResp{status: 204},
	}
	var out, errOut bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader(""), &out, &errOut, "my-plugin", true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), `"pluginInstanceId":"pi-1"`) {
		t.Fatalf("expected deleted instance JSON on stdout, got: %s", out.String())
	}
	if strings.Contains(out.String(), "Deleted plugin instance") {
		t.Fatalf("--json stdout should be pure JSON, got: %s", out.String())
	}
}

func TestRunDelete_DeleteErrorMapping(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: resolvedAliasBody}},
		deleteResp: stubResp{status: 403, body: `{"message":"Forbidden"}`},
	}
	var out bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader("y\n"), &out, &out, "my-plugin", false, false)
	if err == nil || !strings.Contains(err.Error(), "not authorized to delete") {
		t.Fatalf("expected delete forbidden mapping, got: %v", err)
	}
}

func TestRunDelete_TransportError(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: resolvedAliasBody}},
		deleteResp: stubResp{err: errors.New("connection refused")},
	}
	var out bytes.Buffer

	err := runDelete(context.Background(), c, strings.NewReader("y\n"), &out, &out, "my-plugin", false, false)
	if err == nil || !strings.Contains(err.Error(), "failed to delete plugin instance") {
		t.Fatalf("expected wrapped transport error, got: %v", err)
	}
}

func TestMapUMSErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"list 400", mapUMSListError(400, []byte(`{"message":"bad"}`)), "invalid request listing"},
		{"list 403", mapUMSListError(403, []byte(`{}`)), "not authorized to list"},
		{"list 404", mapUMSListError(404, []byte(`{}`)), "not enabled"},
		{"list 500", mapUMSListError(500, []byte(`{"message":"boom"}`)), "status 500"},
		{"lookup 403", mapUMSLookupError(403, []byte(`{}`), "x"), "not authorized to read"},
		{"lookup 404", mapUMSLookupError(404, []byte(`{}`), "x"), "not found"},
		{"delete 403", mapUMSDeleteError(403, []byte(`{}`), "x"), "not authorized to delete"},
		{"delete 404", mapUMSDeleteError(404, []byte(`{}`), "x"), "not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil || !strings.Contains(tt.err.Error(), tt.want) {
				t.Fatalf("expected %q, got: %v", tt.want, tt.err)
			}
		})
	}
}
