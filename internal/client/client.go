// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

type Client interface {
	Get(ctx context.Context, url string) (*http.Response, error)
	Delete(ctx context.Context, url string, params map[string]string) (*http.Response, error)
	Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
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

func (c *SpClient) Get(ctx context.Context, url string) (*http.Response, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	baseUrl := config.GetBaseUrl()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseUrl+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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

func (c *SpClient) Delete(ctx context.Context, url string, params map[string]string) (*http.Response, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	baseUrl := config.GetBaseUrl()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, baseUrl+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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

func (c *SpClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	baseUrl := config.GetBaseUrl()

	if strings.HasPrefix(url, "http") {
		baseUrl = url
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseUrl, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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

func (c *SpClient) Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	baseUrl := config.GetBaseUrl()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, baseUrl+url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

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
