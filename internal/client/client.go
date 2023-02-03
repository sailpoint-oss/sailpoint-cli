// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	clierrors "github.com/sailpoint-oss/sailpoint-cli/internal/errors"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/viper"
)

type Client interface {
	Get(ctx context.Context, url string) (*http.Response, error)
	Delete(ctx context.Context, url string, params map[string]string) (*http.Response, error)
	Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
}

// SpClient provides access to SP APIs.
type SpClient struct {
	cfg         types.CLIConfig
	client      *http.Client
	accessToken string
}

func NewSpClient(cfg types.CLIConfig) Client {
	return &SpClient{
		cfg:    cfg,
		client: &http.Client{},
	}
}

func (c *SpClient) Get(ctx context.Context, url string) (*http.Response, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, err
	}

	baseUrl, err := c.cfg.GetBaseUrl()
	if err != nil {
		return nil, err
	}

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

	baseUrl, err := c.cfg.GetBaseUrl()
	if err != nil {
		return nil, err
	}

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

	baseUrl, err := c.cfg.GetBaseUrl()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseUrl+url, body)
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

	baseUrl, err := c.cfg.GetBaseUrl()
	if err != nil {
		return nil, err
	}

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
	err := c.cfg.Validate()
	if err != nil {
		return err
	}

	if c.accessToken != "" {
		return nil
	}

	var cachedTokenExpiry time.Time
	switch c.cfg.GetAuthType() {
	case "pat":
		cachedTokenExpiry = viper.GetTime("pat.token.expiry")
		if cachedTokenExpiry.After(time.Now()) {
			c.accessToken = viper.GetString("pat.token.accesstoken")
		} else {
			return clierrors.ErrAccessTokenExpired
		}
	case "oauth":
		cachedTokenExpiry = viper.GetTime("oauth.token.expiry")
		if cachedTokenExpiry.After(time.Now()) {
			c.accessToken = viper.GetString("oauth.token.accesstoken")
		} else {
			return clierrors.ErrAccessTokenExpired
		}
	default:
		return errors.New("invalid authtype configured")

	}

	if c.accessToken != "" {
		return fmt.Errorf("no token present")
	}

	return nil

}
