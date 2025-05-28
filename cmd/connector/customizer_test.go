package connector

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// stubClient implements client.Client but should never be called for parent usage
type stubClient struct{}

func (s *stubClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	panic("Get should not be called")
}
func (s *stubClient) Post(ctx context.Context, url, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	panic("Post should not be called")
}
func (s *stubClient) Put(ctx context.Context, url, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	panic("Put should not be called")
}
func (s *stubClient) Delete(ctx context.Context, url string, params, headers map[string]string) (*http.Response, error) {
	panic("Delete should not be called")
}
func (s *stubClient) Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	panic("Patch should not be called")
}

func TestNewConnCustomizersCmd_IncludesAllSubcommands(t *testing.T) {
	client := &stubClient{}
	cmd := newConnCustomizersCmd(client)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	out := buf.String()
	expected := []string{
		"init", "list", "create", "get", "update", "delete", "upload", "link", "unlink",
	}
	for _, sub := range expected {
		if !strings.Contains(out, sub) {
			t.Errorf("usage missing subcommand %q; got:\n%s", sub, out)
		}
	}
}
