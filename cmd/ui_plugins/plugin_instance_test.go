package ui_plugins

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// stubResp is a canned HTTP response for the sequenced test client.
type stubResp struct {
	status int
	body   string
	err    error
}

// seqClient is a test double for client.Client that returns queued Get responses in
// order (for pagination and lookups), a single canned Delete response, and a single
// canned Patch response. It records the URLs, bodies, and headers it was called with.
// Post/Put are unused here.
type seqClient struct {
	getQueue   []stubResp
	getURLs    []string
	getHeaders []map[string]string

	deleteResp    stubResp
	deleteCalls   int
	deleteURL     string
	deleteHeaders map[string]string

	patchResp    stubResp
	patchCalls   int
	patchURL     string
	patchBody    []byte
	patchHeaders map[string]string
}

func (s *seqClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	s.getURLs = append(s.getURLs, url)
	s.getHeaders = append(s.getHeaders, headers)
	if len(s.getQueue) == 0 {
		return nil, errors.New("seqClient: no queued Get response")
	}
	r := s.getQueue[0]
	s.getQueue = s.getQueue[1:]
	if r.err != nil {
		return nil, r.err
	}
	return &http.Response{StatusCode: r.status, Body: io.NopCloser(strings.NewReader(r.body))}, nil
}

func (s *seqClient) Delete(ctx context.Context, url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	s.deleteCalls++
	s.deleteURL = url
	s.deleteHeaders = headers
	if s.deleteResp.err != nil {
		return nil, s.deleteResp.err
	}
	return &http.Response{StatusCode: s.deleteResp.status, Body: io.NopCloser(strings.NewReader(s.deleteResp.body))}, nil
}

func (s *seqClient) Post(ctx context.Context, url, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return nil, errors.New("unused")
}
func (s *seqClient) Put(ctx context.Context, url, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return nil, errors.New("unused")
}
func (s *seqClient) Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	s.patchCalls++
	s.patchURL = url
	s.patchHeaders = headers
	if body != nil {
		s.patchBody, _ = io.ReadAll(body)
	}
	if s.patchResp.err != nil {
		return nil, s.patchResp.err
	}
	return &http.Response{StatusCode: s.patchResp.status, Body: io.NopCloser(strings.NewReader(s.patchResp.body))}, nil
}

func TestListAndDeleteSendExperimentalHeader(t *testing.T) {
	// list
	lc := &seqClient{getQueue: []stubResp{{status: 200, body: `[]`}}}
	if _, err := listPluginInstances(context.Background(), lc); err != nil {
		t.Fatalf("list: unexpected error: %v", err)
	}
	if got := lc.getHeaders[0][experimentalHeader]; got != "true" {
		t.Fatalf("list request missing %s header, got headers: %v", experimentalHeader, lc.getHeaders[0])
	}

	// delete (both the resolve GET and the DELETE must carry the header)
	dc := &seqClient{
		getQueue:   []stubResp{{status: 200, body: `{"pluginInstanceId":"pi-1","alias":"a"}`}},
		deleteResp: stubResp{status: 204},
	}
	var sink bytes.Buffer
	if err := runDelete(context.Background(), dc, strings.NewReader(""), &sink, &sink, "a", true, false); err != nil {
		t.Fatalf("delete: unexpected error: %v", err)
	}
	if got := dc.getHeaders[0][experimentalHeader]; got != "true" {
		t.Fatalf("resolve request missing %s header, got: %v", experimentalHeader, dc.getHeaders[0])
	}
	if got := dc.deleteHeaders[experimentalHeader]; got != "true" {
		t.Fatalf("delete request missing %s header, got: %v", experimentalHeader, dc.deleteHeaders)
	}
}

func TestLooksLikeUUID(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"2c918085-7a1e-1b2c-817a-1e1b2c000000", true},
		{"  2c918085-7a1e-1b2c-817a-1e1b2c000000  ", true},
		{"access-request-plugin", false},
		{"", false},
		{"2c918085-7a1e-1b2c-817a", false},
		{"deadbeef-dead-beef-dead-beefdeadbeef", true}, // UUID-shaped alias: accepted edge case
	}
	for _, tt := range tests {
		if got := looksLikeUUID(tt.in); got != tt.want {
			t.Fatalf("looksLikeUUID(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestTruncateForTable(t *testing.T) {
	tests := []struct {
		in   string
		max  int
		want string
	}{
		{"short", 40, "short"},
		{"exactly-ten", 11, "exactly-ten"}, // len == max, unchanged
		{"this-is-a-very-long-alias-value", 10, "this-is-a…"},
		{"café-plugin-with-accents-that-runs-long", 10, "café-plug…"}, // rune-safe
		{"x", 1, "x"},
		{"xy", 1, "…"},
	}
	for _, tt := range tests {
		if got := truncateForTable(tt.in, tt.max); got != tt.want {
			t.Fatalf("truncateForTable(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
		}
	}
	// The multibyte case must not split a rune: result stays valid UTF-8.
	if got := truncateForTable("héllo-wörld-café", 8); len([]rune(got)) != 8 {
		t.Fatalf("expected 8 runes, got %d (%q)", len([]rune(got)), got)
	}
}

func TestLocalizedName(t *testing.T) {
	tests := []struct {
		name map[string]string
		want string
	}{
		{map[string]string{"en": "English", "fr": "French"}, "English"},
		{map[string]string{"en-US": "US English"}, "US English"},
		{map[string]string{"fr": "French", "de": "German"}, "German"}, // first sorted key
		{map[string]string{}, ""},
		{nil, ""},
	}
	for _, tt := range tests {
		if got := localizedName(tt.name); got != tt.want {
			t.Fatalf("localizedName(%v) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestListPluginInstances_SinglePage(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 200, body: `[{"pluginInstanceId":"a"},{"pluginInstanceId":"b"}]`}}}

	items, err := listPluginInstances(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if len(c.getURLs) != 1 {
		t.Fatalf("expected 1 Get, got %d", len(c.getURLs))
	}
	if !strings.Contains(c.getURLs[0], "limit=250") || !strings.Contains(c.getURLs[0], "offset=0") {
		t.Fatalf("unexpected list URL: %s", c.getURLs[0])
	}
}

func TestListPluginInstances_Paginates(t *testing.T) {
	full := "[" + strings.Repeat(`{"pluginInstanceId":"x"},`, pluginInstancesPageSize-1) + `{"pluginInstanceId":"x"}]`
	c := &seqClient{getQueue: []stubResp{
		{status: 200, body: full},                         // exactly 250 -> fetch again
		{status: 200, body: `[{"pluginInstanceId":"y"}]`}, // short page -> stop
	}}

	items, err := listPluginInstances(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != pluginInstancesPageSize+1 {
		t.Fatalf("expected %d items, got %d", pluginInstancesPageSize+1, len(items))
	}
	if len(c.getURLs) != 2 {
		t.Fatalf("expected 2 Gets, got %d", len(c.getURLs))
	}
	if !strings.Contains(c.getURLs[1], "offset=250") {
		t.Fatalf("expected second page offset=250, got: %s", c.getURLs[1])
	}
}

func TestListPluginInstances_ErrorMapping(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 403, body: `{"message":"Forbidden"}`}}}
	_, err := listPluginInstances(context.Background(), c)
	if err == nil || !strings.Contains(err.Error(), "not authorized to list") {
		t.Fatalf("expected forbidden mapping, got: %v", err)
	}
}

func TestResolveDeleteTarget_UUIDUsesGetByID(t *testing.T) {
	id := "2c918085-7a1e-1b2c-817a-1e1b2c000000"
	c := &seqClient{getQueue: []stubResp{{status: 200, body: `{"pluginInstanceId":"` + id + `","alias":"my-plugin"}`}}}

	inst, _, err := resolvePluginTarget(context.Background(), c, id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst.PluginInstanceID != id {
		t.Fatalf("expected id %s, got %s", id, inst.PluginInstanceID)
	}
	if !strings.HasSuffix(c.getURLs[0], "/"+id) {
		t.Fatalf("expected get-by-id URL, got: %s", c.getURLs[0])
	}
}

func TestResolveDeleteTarget_AliasUsesResolve(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 200, body: `{"pluginInstanceId":"pi-1","alias":"my-plugin"}`}}}

	inst, _, err := resolvePluginTarget(context.Background(), c, "my-plugin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst.PluginInstanceID != "pi-1" {
		t.Fatalf("expected pi-1, got %s", inst.PluginInstanceID)
	}
	if !strings.Contains(c.getURLs[0], "resolve-alias") || !strings.Contains(c.getURLs[0], "alias=my-plugin") {
		t.Fatalf("expected resolve-alias URL, got: %s", c.getURLs[0])
	}
}

func TestResolveDeleteTarget_NotFound(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 404, body: `{"message":"Not Found"}`}}}
	_, _, err := resolvePluginTarget(context.Background(), c, "missing-plugin")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got: %v", err)
	}
}

func TestResolveDeleteTarget_AmbiguousAlias(t *testing.T) {
	body := `{"message":"Alias resolves to multiple plugin instances","conflicts":[{"pluginInstanceId":"pi-1"},{"pluginInstanceId":"pi-2"}]}`
	c := &seqClient{getQueue: []stubResp{{status: 409, body: body}}}

	_, _, err := resolvePluginTarget(context.Background(), c, "dup-plugin")
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	if !strings.Contains(err.Error(), "pi-1") || !strings.Contains(err.Error(), "pi-2") {
		t.Fatalf("expected conflicting IDs in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "specific plugin ID") {
		t.Fatalf("expected guidance to use a plugin ID, got: %v", err)
	}
}

func TestPluginInstanceLabel(t *testing.T) {
	withAlias := &pluginInstance{PluginInstanceID: "pi-1", Alias: "my-plugin"}
	if got := pluginInstanceLabel(withAlias); got != "pi-1 (alias: my-plugin)" {
		t.Fatalf("with alias: got %q", got)
	}

	noAlias := &pluginInstance{PluginInstanceID: "pi-1"}
	if got := pluginInstanceLabel(noAlias); got != "pi-1" {
		t.Fatalf("without alias: got %q", got)
	}
}

func TestSetPluginInstanceState_DisableSuccess(t *testing.T) {
	c := &seqClient{patchResp: stubResp{status: 200, body: `{"pluginInstanceId":"pi-1","state":"DISABLED"}`}}
	inst := &pluginInstance{PluginInstanceID: "pi-1", Alias: "my-plugin"}

	body, err := setPluginInstanceState(context.Background(), c, inst, stateDisabled)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 1 {
		t.Fatalf("expected 1 Patch, got %d", c.patchCalls)
	}
	if !strings.HasSuffix(c.patchURL, "/pi-1") {
		t.Fatalf("expected PATCH on the instance id, got: %s", c.patchURL)
	}
	// The Content-Type header is required: without it UMS silently ignores the
	// body, bumps `modified`, and leaves state unchanged. Guard against regressing it.
	if got := c.patchHeaders["Content-Type"]; got != "application/json" {
		t.Fatalf("PATCH must set Content-Type application/json, got %q (headers: %v)", got, c.patchHeaders)
	}
	if got := c.patchHeaders[experimentalHeader]; got != "true" {
		t.Fatalf("PATCH missing %s header, got: %v", experimentalHeader, c.patchHeaders)
	}
	if got := strings.TrimSpace(string(c.patchBody)); got != `{"state":"DISABLED"}` {
		t.Fatalf("unexpected PATCH body: %s", got)
	}
	if !strings.Contains(string(body), `"state":"DISABLED"`) {
		t.Fatalf("expected the PATCH response body returned, got: %s", body)
	}
}

func TestSetPluginInstanceState_EnableBody(t *testing.T) {
	c := &seqClient{patchResp: stubResp{status: 200, body: `{}`}}
	inst := &pluginInstance{PluginInstanceID: "pi-1"}

	if _, err := setPluginInstanceState(context.Background(), c, inst, stateEnabled); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(string(c.patchBody)); got != `{"state":"ENABLED"}` {
		t.Fatalf("unexpected PATCH body: %s", got)
	}
}

func TestSetPluginInstanceState_ErrorMapping(t *testing.T) {
	c := &seqClient{patchResp: stubResp{status: 403, body: `{"message":"Forbidden"}`}}
	inst := &pluginInstance{PluginInstanceID: "pi-1", Alias: "my-plugin"}

	_, err := setPluginInstanceState(context.Background(), c, inst, stateDisabled)
	if err == nil || !strings.Contains(err.Error(), "not authorized to update") {
		t.Fatalf("expected update forbidden mapping, got: %v", err)
	}
}

func TestSetPluginInstanceState_TransportError(t *testing.T) {
	c := &seqClient{patchResp: stubResp{err: errors.New("connection refused")}}
	inst := &pluginInstance{PluginInstanceID: "pi-1"}

	_, err := setPluginInstanceState(context.Background(), c, inst, stateDisabled)
	if err == nil || !strings.Contains(err.Error(), "failed to update plugin instance state") {
		t.Fatalf("expected wrapped transport error, got: %v", err)
	}
}
