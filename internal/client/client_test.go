// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package client

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

// recordingTransport captures the request it is given and returns an empty
// response without opening a network connection.
type recordingTransport struct {
	req *http.Request
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.req = req
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func testClient(transport http.RoundTripper) *SpClient {
	return &SpClient{
		cfg:         config.CLIConfig{},
		client:      &http.Client{Transport: transport},
		accessToken: "test-token-not-a-real-credential",
	}
}

func TestUrlBuilder(t *testing.T) {
	originalURL := os.Getenv("SAIL_BASE_URL")
	os.Setenv("SAIL_BASE_URL", "https://example.com")
	defer os.Setenv("SAIL_BASE_URL", originalURL)

	spClient := testClient(nil)

	url, sendAuth, err := spClient.resolveUrl("/url/path")
	if err != nil {
		t.Fatalf("expected no error, but got: %s", err)
	}
	if url != "https://example.com/url/path" {
		t.Fatalf("expected url to be: \"https://example.com/url/path\", but got: %s", url)
	}
	if !sendAuth {
		t.Fatal("expected credentials to be sent to the configured tenant")
	}

	// The same origin spelled out in full stays accepted, including the
	// implicit https port.
	url, sendAuth, err = spClient.resolveUrl("https://example.com:443/url/path")
	if err != nil {
		t.Fatalf("expected no error, but got: %s", err)
	}
	if url != "https://example.com:443/url/path" {
		t.Fatalf("expected the same origin url to be kept, but got: %s", url)
	}
	if !sendAuth {
		t.Fatal("expected credentials to be sent to the configured tenant")
	}

	// Local connector instances stay reachable, but without credentials.
	url, sendAuth, err = spClient.resolveUrl("http://localhost:3000")
	if err != nil {
		t.Fatalf("expected no error, but got: %s", err)
	}
	if url != "http://localhost:3000" {
		t.Fatalf("expected url to be: \"http://localhost:3000\", but got: %s", url)
	}
	if sendAuth {
		t.Fatal("expected no credentials to be sent to a loopback endpoint")
	}
}

func TestUrlBuilderRejectsOtherOrigins(t *testing.T) {
	originalURL := os.Getenv("SAIL_BASE_URL")
	os.Setenv("SAIL_BASE_URL", "https://example.com")
	defer os.Setenv("SAIL_BASE_URL", originalURL)

	spClient := testClient(nil)

	cases := []string{
		"https://example.invalid/url/path",     // different host
		"http://example.com/url/path",          // different scheme
		"https://example.com:8443/url/path",    // different port
		"https://user:pw@example.com/url/path", // embedded credentials
		"//example.invalid/url/path",           // protocol relative
		"https://example.com.example.invalid/", // host prefix match only
	}

	for _, rawURL := range cases {
		if _, _, err := spClient.resolveUrl(rawURL); err == nil {
			t.Fatalf("expected %q to be rejected, but it was accepted", rawURL)
		}
	}
}

func TestUrlBuilderRequiresConfiguredTenant(t *testing.T) {
	originalURL := os.Getenv("SAIL_BASE_URL")
	os.Setenv("SAIL_BASE_URL", "")
	defer os.Setenv("SAIL_BASE_URL", originalURL)

	spClient := testClient(nil)

	if _, _, err := spClient.resolveUrl("/url/path"); err == nil {
		t.Fatal("expected a relative path to be rejected with no tenant configured")
	}
}

// TestRequestsToOtherOriginsCarryNoToken is the regression test for
// DEVREL-3142: an endpoint override must not be able to send the tenant
// access token to a host the user did not configure.
func TestRequestsToOtherOriginsCarryNoToken(t *testing.T) {
	originalURL := os.Getenv("SAIL_BASE_URL")
	os.Setenv("SAIL_BASE_URL", "https://example.com")
	defer os.Setenv("SAIL_BASE_URL", originalURL)

	transport := &recordingTransport{}
	spClient := testClient(transport)

	if _, err := spClient.Get(context.Background(), "https://example.invalid/v3/accounts", nil); err == nil {
		t.Fatal("expected the request to another origin to fail")
	}
	if transport.req != nil {
		t.Fatalf("expected no request to be sent, but one went to: %s", transport.req.URL)
	}

	if _, err := spClient.Get(context.Background(), "http://127.0.0.1:3000/invoke", nil); err != nil {
		t.Fatalf("expected the loopback request to succeed, but got: %s", err)
	}
	if transport.req == nil {
		t.Fatal("expected the loopback request to be sent")
	}
	if got := transport.req.Header.Get("Authorization"); got != "" {
		t.Fatalf("expected no authorization header on the loopback request, but got: %s", got)
	}
}
