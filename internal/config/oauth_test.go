package config

import (
	"encoding/json"
	"testing"
)

func TestConfirmationCodeFromID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "uuid uses last eight characters",
			id:   "12345678-90ab-cdef-1234-567890abcdef",
			want: "90AB-CDEF",
		},
		{
			name: "short id is uppercased",
			id:   "abc123",
			want: "ABC123",
		},
		{
			name: "whitespace is ignored",
			id:   "  12345678-90ab-cdef-1234-567890abcdef  ",
			want: "90AB-CDEF",
		},
		{
			name: "empty id stays empty",
			id:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := confirmationCodeFromID(tt.id); got != tt.want {
				t.Fatalf("confirmationCodeFromID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestAuthResponseIncludesPickupSecret(t *testing.T) {
	raw := []byte(`{
		"authURL": "https://example.com/auth",
		"id": "12345678-90ab-cdef-1234-567890abcdef",
		"baseURL": "https://example.api.identitynow.com",
		"pickupSecret": "secret-value",
		"ttl": 123
	}`)

	var response AuthResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.PickupSecret != "secret-value" {
		t.Fatalf("PickupSecret = %q, want %q", response.PickupSecret, "secret-value")
	}
}

func TestNewOAuthTokenRequestUsesPickupSecretBearer(t *testing.T) {
	req, err := newOAuthTokenRequest("https://example.com/auth/token", "session-id", "secret-value")
	if err != nil {
		t.Fatalf("newOAuthTokenRequest() error = %v", err)
	}

	if got, want := req.Method, "GET"; got != want {
		t.Fatalf("Method = %q, want %q", got, want)
	}
	if got, want := req.URL.String(), "https://example.com/auth/token/session-id"; got != want {
		t.Fatalf("URL = %q, want %q", got, want)
	}
	if got, want := req.Header.Get("Authorization"), "Bearer secret-value"; got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}
