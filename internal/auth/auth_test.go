package auth

import (
	"strings"
	"testing"
)

func TestValidateAuthRejectsUnknownAuthType(t *testing.T) {
	err := ValidateAuth("saml", "default", "https://tenant.api.identitynow.com", "https://tenant.identitynow.com")
	if err == nil {
		t.Fatal("expected invalid auth type to fail")
	}
	if !strings.Contains(err.Error(), "configuration invalid") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAuthOAuthRequiresBaseURLAndTenantURL(t *testing.T) {
	err := ValidateAuth("oauth", "default", "", "")
	if err == nil {
		t.Fatal("expected missing OAuth URLs to fail")
	}
	if !strings.Contains(err.Error(), "configuration invalid") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTokenRejectsInvalidAuthType(t *testing.T) {
	_, err := GetToken("bad", "default", "https://tenant.api.identitynow.com", "https://tenant.identitynow.com", "https://tenant.api.identitynow.com/oauth/token", nil)
	if err == nil {
		t.Fatal("expected invalid auth type to fail")
	}
	if !strings.Contains(err.Error(), "configuration invalid") {
		t.Fatalf("unexpected error: %v", err)
	}
}
