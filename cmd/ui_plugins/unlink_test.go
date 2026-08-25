package ui_plugins

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRunUnlink_Success(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: linkResolvedBody}},
		deleteResp: stubResp{status: 204},
	}
	var out bytes.Buffer

	err := runUnlink(context.Background(), c, tempManifestPath(t, testManifestJSON), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.deleteCalls != 1 {
		t.Fatalf("expected exactly one Delete, got %d", c.deleteCalls)
	}
	wantURL := pluginInstancesEndpoint + "/pi-123/link"
	if c.deleteURL != wantURL {
		t.Fatalf("DELETE url = %s, want %s", c.deleteURL, wantURL)
	}
	if c.deleteHeaders[experimentalHeader] != "true" {
		t.Fatalf("expected %s header on unlink DELETE, got: %v", experimentalHeader, c.deleteHeaders)
	}
	if got := out.String(); !strings.Contains(got, "Removed the local dev link") || !strings.Contains(got, "access-request-plugin") {
		t.Fatalf("expected confirmation with alias, got: %s", got)
	}
}

// A 204 is returned whether or not a mapping existed (UMS treats unlink as
// idempotent), so the no-op case is indistinguishable from a real removal and
// must still report success.
func TestRunUnlink_NoOpStillConfirms(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: linkResolvedBody}},
		deleteResp: stubResp{status: 204},
	}
	var out bytes.Buffer

	err := runUnlink(context.Background(), c, tempManifestPath(t, testManifestJSON), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Removed the local dev link") {
		t.Fatalf("expected a success confirmation for a no-op unlink, got: %s", out.String())
	}
}

func TestRunUnlink_InvalidManifestSkipsClient(t *testing.T) {
	c := &seqClient{}

	err := runUnlink(context.Background(), c, tempManifestPath(t, `{"version":1}`), io.Discard)
	if err == nil {
		t.Fatal("expected an error for an invalid manifest")
	}
	if len(c.getURLs) != 0 || c.deleteCalls != 0 {
		t.Fatalf("expected no client calls, got get=%d delete=%d", len(c.getURLs), c.deleteCalls)
	}
}

func TestRunUnlink_AliasResolutionFailureSkipsDelete(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 404, body: `{"message":"not found"}`}}}

	err := runUnlink(context.Background(), c, tempManifestPath(t, testManifestJSON), io.Discard)
	if err == nil {
		t.Fatal("expected an error when alias resolution fails")
	}
	if c.deleteCalls != 0 {
		t.Fatalf("expected no Delete after alias resolution failure, got %d", c.deleteCalls)
	}
}

func TestRunUnlink_ForbiddenMapped(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: linkResolvedBody}},
		deleteResp: stubResp{status: 403, body: `{"message":"Forbidden"}`},
	}
	var out bytes.Buffer

	err := runUnlink(context.Background(), c, tempManifestPath(t, testManifestJSON), &out)
	if err == nil || !strings.Contains(err.Error(), "not authorized to unlink") {
		t.Fatalf("expected forbidden mapping, got: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no confirmation on error, got: %s", out.String())
	}
}

func TestRunUnlink_NotFoundMapped(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: linkResolvedBody}},
		deleteResp: stubResp{status: 404, body: `{"message":"Not Found"}`},
	}

	err := runUnlink(context.Background(), c, tempManifestPath(t, testManifestJSON), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found mapping, got: %v", err)
	}
}

func TestRunUnlink_TransportError(t *testing.T) {
	c := &seqClient{
		getQueue:   []stubResp{{status: 200, body: linkResolvedBody}},
		deleteResp: stubResp{err: errors.New("connection refused")},
	}

	err := runUnlink(context.Background(), c, tempManifestPath(t, testManifestJSON), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "failed to remove link") {
		t.Fatalf("expected wrapped transport error, got: %v", err)
	}
}

func TestMapUMSUnlinkError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{"forbidden", 403, "not authorized to unlink"},
		{"not found", 404, "not found"},
		{"unexpected", 500, "status 500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapUMSUnlinkError(tt.status, []byte(`{"message":"boom"}`), "my-plugin")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q, got: %v", tt.want, err)
			}
		})
	}
}
