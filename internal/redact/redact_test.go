package redact

import (
	"strings"
	"testing"
)

func TestStringRedactsHeadersAndKeyValueSecrets(t *testing.T) {
	input := "Authorization: Bearer secret-token\nclient_secret=abc123&next=true\npassword=\"hunter2\""
	output := String(input)

	for _, secret := range []string{"secret-token", "abc123", "hunter2"} {
		if strings.Contains(output, secret) {
			t.Fatalf("expected %q to be redacted from %q", secret, output)
		}
	}
	if strings.Count(output, replacement) < 3 {
		t.Fatalf("expected redacted output, got %q", output)
	}
}

func TestValueRedactsNestedSensitiveKeys(t *testing.T) {
	value := map[string]any{
		"name": "visible",
		"nested": map[string]any{
			"accessToken": "secret",
		},
		"items": []any{
			map[string]any{"clientSecret": "secret2"},
		},
	}

	redacted := Value(value).(map[string]any)
	if redacted["name"] != "visible" {
		t.Fatalf("expected non-sensitive value to remain visible")
	}
	nested := redacted["nested"].(map[string]any)
	if nested["accessToken"] != replacement {
		t.Fatalf("expected nested access token to be redacted")
	}
	items := redacted["items"].([]any)
	item := items[0].(map[string]any)
	if item["clientSecret"] != replacement {
		t.Fatalf("expected nested client secret to be redacted")
	}
}
