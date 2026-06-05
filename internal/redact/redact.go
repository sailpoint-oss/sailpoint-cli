package redact

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const replacement = "[REDACTED]"

var sensitiveKeyFragments = []string{
	"authorization",
	"access_token",
	"accesstoken",
	"refresh_token",
	"refreshtoken",
	"pickupsecret",
	"clientsecret",
	"client_secret",
	"secret",
	"password",
	"token",
	"credential",
}

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?im)^(Authorization:\s*Bearer\s+).+$`),
	regexp.MustCompile(`(?im)^(Authorization:\s*Basic\s+).+$`),
	regexp.MustCompile(`(?i)("?[A-Za-z0-9_.-]*(?:token|secret|password|credential)[A-Za-z0-9_.-]*"?\s*[:=]\s*")([^"]+)(")`),
	regexp.MustCompile(`(?i)((?:access_token|refresh_token|client_secret|pickupSecret|password|secret|token)=)([^&\s]+)`),
}

func String(value string) string {
	redacted := value
	for _, pattern := range sensitivePatterns {
		switch pattern.NumSubexp() {
		case 1:
			redacted = pattern.ReplaceAllString(redacted, `${1}`+replacement)
		case 3:
			redacted = pattern.ReplaceAllString(redacted, `${1}`+replacement+`${3}`)
		default:
			redacted = pattern.ReplaceAllString(redacted, `${1}`+replacement)
		}
	}
	return redacted
}

func Bytes(value []byte) string {
	return String(string(value))
}

func JSONBytes(value []byte) []byte {
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return []byte(String(string(value)))
	}

	redacted := Value(decoded)
	data, err := json.MarshalIndent(redacted, "", "  ")
	if err != nil {
		return []byte(fmt.Sprintf("%v", redacted))
	}
	return data
}

func Value(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, val := range typed {
			if IsSensitiveKey(key) {
				out[key] = replacement
				continue
			}
			out[key] = Value(val)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, val := range typed {
			out[i] = Value(val)
		}
		return out
	case string:
		return String(typed)
	default:
		return typed
	}
}

func IsSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	for _, fragment := range sensitiveKeyFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}
