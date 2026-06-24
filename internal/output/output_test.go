package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestWriteStructuredJSON(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set("output", "json")

	var buf bytes.Buffer
	if err := WriteStructured(&buf, map[string]any{"name": "demo"}); err != nil {
		t.Fatalf("WriteStructured returned error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("expected valid JSON, got %q: %v", buf.String(), err)
	}
	if decoded["name"] != "demo" {
		t.Fatalf("unexpected JSON payload: %#v", decoded)
	}
}

func TestWriteStructuredYAML(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set("output", "yaml")

	var buf bytes.Buffer
	if err := WriteStructured(&buf, map[string]any{"name": "demo"}); err != nil {
		t.Fatalf("WriteStructured returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "name: demo") {
		t.Fatalf("expected YAML payload, got %q", buf.String())
	}
}

func TestCurrentFormatFallsBackToTable(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set("output", "text")

	if got := CurrentFormat(); got != FormatTable {
		t.Fatalf("expected unsupported output format to fall back to table, got %q", got)
	}
}
