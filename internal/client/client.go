// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

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
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getUrl(url), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.getUrl(url), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.getUrl(url), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.getUrl(url), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.getUrl(url), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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

// getUrl constructs the url to call out while supporting url overwrites if full url is provided
func (s *SpClient) getUrl(path string) string {

	u, err := url.Parse(path)
	if err != nil {
		// keep the url building process today if parsing fails
		return config.GetBaseUrl() + path
	}

	if u.Host != "" && u.Scheme != "" {
		return path
	}

	return config.GetBaseUrl() + path
}
