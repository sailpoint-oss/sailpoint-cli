// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Client interface {
	Get(ctx context.Context, url string) (*http.Response, error)
	Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	VerifyToken(ctx context.Context, tokenUrl, clientID, secret string) error
}

// SpClient provides access to SP APIs.
type SpClient struct {
	cfg         SpClientConfig
	client      *http.Client
	accessToken string
}

type SpClientConfig struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Debug        bool
}

func (c *SpClientConfig) Validate() error {
	if c.TokenURL == "" {
		return fmt.Errorf("Missing TokenURL configuration value")
	}
	if c.ClientID == "" {
		return fmt.Errorf("Missing ClientID configuration value")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("Missing ClientSecret configuration value")
	}
	return nil
}

func NewSpClient(cfg SpClientConfig) Client {
	return &SpClient{
		cfg:    cfg,
		client: &http.Client{},
	}
}

func (c *SpClient) Get(ctx context.Context, url string) (*http.Response, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

func (c *SpClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
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
	if c.cfg.Debug {
		dbg, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(dbg))
	}
	return resp, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c *SpClient) ensureAccessToken(ctx context.Context) error {
	err := c.cfg.Validate()
	if err != nil {
		return err
	}

	if c.accessToken != "" {
		return nil
	}

	uri, err := url.Parse(c.cfg.TokenURL)
	if err != nil {
		return err
	}

	query := &url.Values{}
	query.Add("grant_type", "client_credentials")
	uri.RawQuery = query.Encode()

	data := &url.Values{}
	data.Add("client_id", c.cfg.ClientID)
	data.Add("client_secret", c.cfg.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to retrieve access token. status %s", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tResponse tokenResponse
	err = json.Unmarshal(raw, &tResponse)
	if err != nil {
		return err
	}

	c.accessToken = tResponse.AccessToken
	return nil
}

func (c *SpClient) VerifyToken(ctx context.Context, tokenUrl, clientID, secret string) error {
	c.cfg.TokenURL = tokenUrl
	c.cfg.ClientID = clientID
	c.cfg.ClientSecret = secret
	return c.ensureAccessToken(ctx)
}
