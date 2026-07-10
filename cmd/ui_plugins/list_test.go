package ui_plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunList_Table(t *testing.T) {
	body := `[
	  {"pluginInstanceId":"pi-2","alias":"zeta-plugin","name":{"en":"Zeta"},"created":"2026-06-01T00:00:00Z"},
	  {"pluginInstanceId":"pi-1","alias":"alpha-plugin","name":{"en":"Alpha"},"created":"2026-05-01T00:00:00Z"}
	]`
	c := &seqClient{getQueue: []stubResp{{status: 200, body: body}}}
	var out bytes.Buffer

	if err := runList(context.Background(), c, &out, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	// tablewriter upper-cases headers; assert on the header labels (upper) and the row data.
	for _, want := range []string{"ALIAS", "ID", "NAME", "CREATED", "alpha-plugin", "zeta-plugin", "pi-1", "Alpha"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected table to contain %q, got:\n%s", want, got)
		}
	}
	// Sorted by alias: alpha before zeta.
	if strings.Index(got, "alpha-plugin") > strings.Index(got, "zeta-plugin") {
		t.Fatalf("expected rows sorted by alias, got:\n%s", got)
	}
}

func TestRunList_Empty(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 200, body: `[]`}}}
	var out bytes.Buffer

	if err := runList(context.Background(), c, &out, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "No plugin instances found.") {
		t.Fatalf("expected friendly empty message, got: %s", out.String())
	}
}

func TestRunList_JSONFidelity(t *testing.T) {
	// Includes a field not modeled by pluginInstance to prove --json is loss-free.
	body := `[{"pluginInstanceId":"pi-1","alias":"a","devOverrides":{"someId":"http://localhost"}}]`
	c := &seqClient{getQueue: []stubResp{{status: 200, body: body}}}
	var out bytes.Buffer

	if err := runList(context.Background(), c, &out, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "devOverrides") || !strings.Contains(out.String(), "localhost") {
		t.Fatalf("expected raw fields preserved in --json, got: %s", out.String())
	}
	// Output must be a valid JSON array.
	var arr []map[string]any
	if err := json.Unmarshal([]byte(out.String()), &arr); err != nil {
		t.Fatalf("expected valid JSON array, got error %v for: %s", err, out.String())
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr))
	}
}

func TestRunList_JSONEmptyIsArray(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: 200, body: `[]`}}}
	var out bytes.Buffer

	if err := runList(context.Background(), c, &out, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out.String()) != "[]" {
		t.Fatalf("expected empty JSON array, got: %s", out.String())
	}
}
