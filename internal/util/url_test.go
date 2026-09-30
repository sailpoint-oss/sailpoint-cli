package util

import "testing"

func TestParseTenantInput(t *testing.T) {
	tests := []struct {
		input     string
		name      string
		tenantURL string
		baseURL   string
	}{
		{"acme", "acme", "https://acme.identitynow.com", "https://acme.api.identitynow.com"},
		{"  acme  ", "acme", "https://acme.identitynow.com", "https://acme.api.identitynow.com"},
		{"https://acme.identitynow.com", "acme", "https://acme.identitynow.com", "https://acme.api.identitynow.com"},
		{"https://devrel-ga-24331.identitynow-demo.com/", "devrel-ga-24331", "https://devrel-ga-24331.identitynow-demo.com", "https://devrel-ga-24331.api.identitynow-demo.com"},
		{"https://devrel-ga-24331.identitynow-demo.com/ui/d/dashboard?x=1", "devrel-ga-24331", "https://devrel-ga-24331.identitynow-demo.com", "https://devrel-ga-24331.api.identitynow-demo.com"},
		{"devrel-ga-24331.identitynow-demo.com", "devrel-ga-24331", "https://devrel-ga-24331.identitynow-demo.com", "https://devrel-ga-24331.api.identitynow-demo.com"},
		{"https://devrel-ga-24331.api.identitynow-demo.com", "devrel-ga-24331", "https://devrel-ga-24331.identitynow-demo.com", "https://devrel-ga-24331.api.identitynow-demo.com"},
		{"HTTPS://Acme.IdentityNow.com", "acme", "https://acme.identitynow.com", "https://acme.api.identitynow.com"},
	}

	for _, tt := range tests {
		name, tenantURL, baseURL, err := ParseTenantInput(tt.input)
		if err != nil {
			t.Errorf("ParseTenantInput(%q) returned error: %v", tt.input, err)
			continue
		}
		if name != tt.name || tenantURL != tt.tenantURL || baseURL != tt.baseURL {
			t.Errorf("ParseTenantInput(%q) = (%q, %q, %q), want (%q, %q, %q)", tt.input, name, tenantURL, baseURL, tt.name, tt.tenantURL, tt.baseURL)
		}
	}
}

func TestParseTenantInputRejectsBadInput(t *testing.T) {
	for _, input := range []string{"", "   ", "http://acme.identitynow.com", "https://", "https://.identitynow.com", "https://localhost"} {
		if _, _, _, err := ParseTenantInput(input); err == nil {
			t.Errorf("ParseTenantInput(%q) returned no error", input)
		}
	}
}
