// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

type Client interface {
	Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error)
	Delete(ctx context.Context, url string, params map[string]string, headers map[string]string) (*http.Response, error)
	Post(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error)
	Put(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error)
	Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error)
}

// SpClient provides access to SP APIs.
type SpClient struct {
	cfg         config.CLIConfig
	client      *http.Client
	accessToken string
}

func NewSpClient(cfg config.CLIConfig) Client {
	return &SpClient{
		cfg:    cfg,
		client: &http.Client{},
	}
}

func (c *SpClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	target, sendAuth, err := c.resolveUrl(url)
	if err != nil {
		return nil, err
	}

	if sendAuth {
		if err := c.ensureAccessToken(ctx); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	if sendAuth {
		req.Header.Add("Authorization", "Bearer "+c.accessToken)
	}

	// Add any additional headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(dbg))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(dbg))
	}
	return resp, nil
}

func (c *SpClient) Delete(ctx context.Context, url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	target, sendAuth, err := c.resolveUrl(url)
	if err != nil {
		return nil, err
	}

	if sendAuth {
		if err := c.ensureAccessToken(ctx); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, target, nil)
	if err != nil {
		return nil, err
	}
	if sendAuth {
		req.Header.Add("Authorization", "Bearer "+c.accessToken)
	}

	// Add any additional headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(dbg))
	}

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(dbg))
	}
	return resp, nil
}

func (c *SpClient) Post(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	target, sendAuth, err := c.resolveUrl(url)
	if err != nil {
		return nil, err
	}

	if sendAuth {
		if err := c.ensureAccessToken(ctx); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	if sendAuth {
		req.Header.Add("Authorization", "Bearer "+c.accessToken)
	}

	// Add any additional headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(dbg))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if c.cfg.Debug {
		dbg, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(dbg))
	}
	return resp, nil
}

func (c *SpClient) Put(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	target, sendAuth, err := c.resolveUrl(url)
	if err != nil {
		return nil, err
	}

	if sendAuth {
		if err := c.ensureAccessToken(ctx); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	if sendAuth {
		req.Header.Add("Authorization", "Bearer "+c.accessToken)
	}

	// Add any additional headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(dbg))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(dbg))
	}

	return resp, nil
}

func (c *SpClient) Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	target, sendAuth, err := c.resolveUrl(url)
	if err != nil {
		return nil, err
	}

	if sendAuth {
		if err := c.ensureAccessToken(ctx); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, target, body)
	if err != nil {
		return nil, err
	}
	if sendAuth {
		req.Header.Add("Authorization", "Bearer "+c.accessToken)
	}

	// Add any additional headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(dbg))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if c.cfg.Debug {
		dbg, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(dbg))
	}

	return resp, nil
}

func (c *SpClient) ensureAccessToken(ctx context.Context) error {
	token, err := config.GetAuthToken()
	if err != nil {
		return err
	}

	c.accessToken = token

	return nil
}

// resolveUrl determines the absolute url for a request and whether the
// configured tenant credentials may be attached to it.
//
// A relative path is resolved against the configured tenant base url. An
// absolute url is only accepted when it points at the same origin as the
// configured tenant, so that an endpoint override cannot send the tenant
// access token to another host. Loopback overrides stay available for local
// connector development and are always called without credentials.
func (c *SpClient) resolveUrl(rawURL string) (string, bool, error) {
	trimmed := strings.TrimSpace(rawURL)

	u, err := url.Parse(trimmed)
	if err != nil {
		return "", false, fmt.Errorf("invalid request url %q: %v", rawURL, err)
	}

	// A relative path always belongs to the configured tenant.
	if u.Scheme == "" && u.Host == "" && u.Opaque == "" {
		base := strings.TrimSuffix(config.GetBaseUrl(), "/")
		if base == "" {
			return "", false, fmt.Errorf("no tenant base url is configured, run \"sail configure\" or set SAIL_BASE_URL")
		}
		return base + trimmed, true, nil
	}

	// Reject anything that is neither a plain relative path nor a complete
	// url, such as the protocol relative form "//example.com/path".
	if u.Scheme == "" || u.Host == "" || u.Opaque != "" {
		return "", false, fmt.Errorf("request url %q must either be a path or a complete url", rawURL)
	}

	if u.User != nil {
		return "", false, fmt.Errorf("request url %q must not include credentials", rawURL)
	}

	// Local connector instances are reached without tenant credentials.
	if isLoopbackHost(u.Hostname()) {
		return u.String(), false, nil
	}

	base, err := url.Parse(strings.TrimSuffix(config.GetBaseUrl(), "/"))
	if err != nil || base.Host == "" {
		return "", false, fmt.Errorf("no valid tenant base url is configured, run \"sail configure\" or set SAIL_BASE_URL")
	}

	if !sameOrigin(base, u) {
		return "", false, fmt.Errorf("refusing to send tenant credentials to %q, which is not the configured tenant %q", origin(u), origin(base))
	}

	return u.String(), true, nil
}

// sameOrigin reports whether two urls share a scheme, host and port.
func sameOrigin(a *url.URL, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) &&
		strings.EqualFold(a.Hostname(), b.Hostname()) &&
		port(a) == port(b)
}

// port returns the explicit port of u, or the default port for its scheme.
func port(u *url.URL) string {
	if p := u.Port(); p != "" {
		return p
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}

// origin returns the scheme, host and port of u for use in error messages.
func origin(u *url.URL) string {
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
}

// isLoopbackHost reports whether host refers to the local machine.
func isLoopbackHost(host string) bool {
	host = strings.ToLower(host)
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
