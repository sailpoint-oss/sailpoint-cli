package ui_plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

// resolveRenameBody is a resolve-alias 200 response for the testManifestJSON
// fixture, carrying a name map so the name-change PATCH path can preserve other
// locales.
const resolveRenameBody = `{"pluginInstanceId":"pi-123","alias":"access-request-plugin","name":{"en-US":"Access Request"},"slots":[{"slotId":"full-page"}]}`

// readManifest reads the manifest file back and returns its manifest subsection
// as a generic map for field-level assertions.
func readManifest(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	m, ok := root["manifest"].(map[string]any)
	if !ok {
		t.Fatalf("manifest section missing in %s", path)
	}
	return m
}

func renameTestDeps(c *seqClient, stdin string, out, errOut *bytes.Buffer) renameDeps {
	return renameDeps{client: c, in: strings.NewReader(stdin), out: out, errOut: errOut}
}

func TestRunRename_NoChange_DoesNotTouchUMSOrFile(t *testing.T) {
	c := &seqClient{}
	var out bytes.Buffer
	path := tempManifestPath(t, testManifestJSON)

	opts := renameOptions{name: "Access Request", nameSet: true, alias: "access-request-plugin", aliasSet: true}
	if err := runRename(context.Background(), renameTestDeps(c, "", &out, &bytes.Buffer{}), path, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.getURLs) != 0 {
		t.Fatalf("expected no UMS calls on a no-op rename, got Gets: %v", c.getURLs)
	}
	if c.patchCalls != 0 {
		t.Fatalf("expected no PATCH on a no-op rename, got %d", c.patchCalls)
	}
	if !strings.Contains(out.String(), "No changes") {
		t.Fatalf("expected a no-change message, got: %s", out.String())
	}
}

func TestRunRename_NameOnly_ExistingPlugin(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: http.StatusOK, body: resolveRenameBody}},
		patchResp: stubResp{status: http.StatusOK, body: `{}`},
	}
	var out bytes.Buffer
	path := tempManifestPath(t, testManifestJSON)

	opts := renameOptions{name: "New Name", nameSet: true}
	if err := runRename(context.Background(), renameTestDeps(c, "", &out, &bytes.Buffer{}), path, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.patchCalls != 1 {
		t.Fatalf("expected one PATCH, got %d", c.patchCalls)
	}
	body := string(c.patchBody)
	if !strings.Contains(body, `"New Name"`) || !strings.Contains(body, "en-US") {
		t.Fatalf("expected name map with new en and preserved en-US, got: %s", body)
	}
	if strings.Contains(body, `"alias"`) {
		t.Fatalf("alias must not be sent when only the name changed, got: %s", body)
	}
	if c.patchHeaders["Content-Type"] != "application/json" {
		t.Fatalf("expected Content-Type application/json, got: %v", c.patchHeaders)
	}

	m := readManifest(t, path)
	name := m["name"].(map[string]any)
	if name["en"] != "New Name" {
		t.Fatalf("expected file name.en updated, got: %v", name)
	}
	if name["en-US"] != "Access Request" {
		t.Fatalf("expected file name.en-US preserved, got: %v", name)
	}
	if m["alias"] != "access-request-plugin" {
		t.Fatalf("expected file alias unchanged, got: %v", m["alias"])
	}
	if strings.Contains(out.String(), "ui-plugins link") {
		t.Fatalf("link next-step should only print on an alias change, got: %s", out.String())
	}
}

func TestRunRename_AliasChange_ExistingPlugin_NormalizesAndLinksHint(t *testing.T) {
	c := &seqClient{
		getQueue: []stubResp{
			{status: http.StatusOK, body: resolveRenameBody}, // resolve current alias
			{status: http.StatusOK, body: ``},                // validate-alias (available)
		},
		patchResp: stubResp{status: http.StatusOK, body: `{}`},
	}
	var out bytes.Buffer
	path := tempManifestPath(t, testManifestJSON)
	withTenantURL(t, "https://tenant.example.com")

	// Mixed-case input must be normalized to lowercase.
	opts := renameOptions{alias: "New-Alias", aliasSet: true}
	if err := runRename(context.Background(), renameTestDeps(c, "", &out, &bytes.Buffer{}), path, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.getURLs) != 2 {
		t.Fatalf("expected resolve + validate-alias Gets, got: %v", c.getURLs)
	}
	if !strings.Contains(c.getURLs[1], "validate-alias") || !strings.Contains(c.getURLs[1], "new-alias") {
		t.Fatalf("expected validate-alias on normalized new alias, got: %s", c.getURLs[1])
	}
	body := string(c.patchBody)
	if !strings.Contains(body, `"alias":"new-alias"`) {
		t.Fatalf("expected lowercased alias in PATCH, got: %s", body)
	}
	if strings.Contains(body, `"name"`) {
		t.Fatalf("name must not be sent when only the alias changed, got: %s", body)
	}
	if m := readManifest(t, path); m["alias"] != "new-alias" {
		t.Fatalf("expected file alias normalized to new-alias, got: %v", m["alias"])
	}
	if !strings.Contains(out.String(), "ui-plugins link") {
		t.Fatalf("expected link next-step after an alias change, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "The plugin can be viewed at:\nhttps://tenant.example.com/ui/plugin/pi-123") {
		t.Fatalf("expected the plugin view URL (as upload renders it), got: %s", out.String())
	}
}

func TestRunRename_AliasChange_NotCreated_WritesFileNoPatch(t *testing.T) {
	c := &seqClient{
		getQueue: []stubResp{
			{status: http.StatusNotFound, body: `{"message":"not found"}`}, // resolve -> 404
			{status: http.StatusOK, body: ``},                              // validate-alias
		},
	}
	var out bytes.Buffer
	path := tempManifestPath(t, testManifestJSON)

	opts := renameOptions{alias: "fresh-alias", aliasSet: true}
	if err := runRename(context.Background(), renameTestDeps(c, "", &out, &bytes.Buffer{}), path, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.patchCalls != 0 {
		t.Fatalf("expected no PATCH when the plugin does not exist, got %d", c.patchCalls)
	}
	if m := readManifest(t, path); m["alias"] != "fresh-alias" {
		t.Fatalf("expected file alias written, got: %v", m["alias"])
	}
	got := out.String()
	if !strings.Contains(got, "not been created") || !strings.Contains(got, "ui-plugins create") {
		t.Fatalf("expected create next-step for an uncreated plugin, got: %s", got)
	}
	if strings.Contains(got, "ui-plugins link") {
		t.Fatalf("link next-step must be omitted when the plugin does not exist, got: %s", got)
	}
}

func TestRunRename_AmbiguousResolve_StopsBeforeWriting(t *testing.T) {
	c := &seqClient{
		getQueue: []stubResp{{status: http.StatusConflict, body: `{"conflicts":[{"pluginInstanceId":"pi-1"},{"pluginInstanceId":"pi-2"}]}`}},
	}
	var out bytes.Buffer
	path := tempManifestPath(t, testManifestJSON)

	opts := renameOptions{alias: "whatever", aliasSet: true}
	err := runRename(context.Background(), renameTestDeps(c, "", &out, &bytes.Buffer{}), path, opts)
	if err == nil {
		t.Fatalf("expected an error on an ambiguous resolve")
	}
	if !strings.Contains(err.Error(), "pi-1") || !strings.Contains(err.Error(), "pi-2") {
		t.Fatalf("expected conflicting ids in the error, got: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatalf("expected no PATCH, got %d", c.patchCalls)
	}
	if m := readManifest(t, path); m["alias"] != "access-request-plugin" {
		t.Fatalf("file must be untouched on an ambiguous resolve, got alias: %v", m["alias"])
	}
}

func TestRunRename_FlaggedAliasTaken_FailsFast(t *testing.T) {
	c := &seqClient{
		getQueue: []stubResp{
			{status: http.StatusOK, body: resolveRenameBody},            // resolve -> exists
			{status: http.StatusConflict, body: `{"message":"in use"}`}, // validate-alias -> taken
		},
	}
	var out bytes.Buffer
	path := tempManifestPath(t, testManifestJSON)

	opts := renameOptions{alias: "taken-alias", aliasSet: true}
	err := runRename(context.Background(), renameTestDeps(c, "", &out, &bytes.Buffer{}), path, opts)
	if err == nil {
		t.Fatalf("expected an error when a flagged alias is already taken")
	}
	if c.patchCalls != 0 {
		t.Fatalf("expected no PATCH on a taken flagged alias, got %d", c.patchCalls)
	}
	if m := readManifest(t, path); m["alias"] != "access-request-plugin" {
		t.Fatalf("file must be untouched on validation failure, got alias: %v", m["alias"])
	}
}

func TestRunRename_InteractiveAliasReprompt(t *testing.T) {
	c := &seqClient{
		getQueue: []stubResp{
			{status: http.StatusOK, body: resolveRenameBody},            // resolve -> exists
			{status: http.StatusConflict, body: `{"message":"in use"}`}, // validate taken-alias
			{status: http.StatusOK, body: ``},                           // validate good-alias
		},
		patchResp: stubResp{status: http.StatusOK, body: `{}`},
	}
	var out bytes.Buffer
	// Name prompt: Enter (keep). Alias prompt: taken-alias, then good-alias.
	stdin := "\ntaken-alias\ngood-alias\n"
	path := tempManifestPath(t, testManifestJSON)

	if err := runRename(context.Background(), renameTestDeps(c, stdin, &out, &bytes.Buffer{}), path, renameOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.patchCalls != 1 {
		t.Fatalf("expected one PATCH after a successful re-prompt, got %d", c.patchCalls)
	}
	if !strings.Contains(string(c.patchBody), `"alias":"good-alias"`) {
		t.Fatalf("expected the re-prompted alias in PATCH, got: %s", c.patchBody)
	}
	if !strings.Contains(out.String(), "please choose a different alias") {
		t.Fatalf("expected a re-prompt message, got: %s", out.String())
	}
	if m := readManifest(t, path); m["alias"] != "good-alias" {
		t.Fatalf("expected file alias set to the re-prompted value, got: %v", m["alias"])
	}
}
