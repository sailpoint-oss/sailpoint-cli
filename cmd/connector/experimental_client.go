// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"context"
	"io"
	"net/http"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
)

// experimentalHeader must be sent on platform-connectors, connector-instances
// and platform-logs v1 requests while those routes are in a preview API
// lifecycle state (public-preview).
const experimentalHeader = "X-SailPoint-Experimental"

// experimentalClient wraps a client.Client and adds the experimental opt-in
// header to every request, so the connector commands do not each need to pass it.
type experimentalClient struct {
	client.Client
}

func newExperimentalClient(c client.Client) client.Client {
	return &experimentalClient{Client: c}
}

// withExperimentalHeader returns a copy of headers with the experimental
// header added. A value the caller already set is kept.
func withExperimentalHeader(headers map[string]string) map[string]string {
	out := make(map[string]string, len(headers)+1)
	for k, v := range headers {
		out[k] = v
	}
	if _, ok := out[experimentalHeader]; !ok {
		out[experimentalHeader] = "true"
	}
	return out
}

func (c *experimentalClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return c.Client.Get(ctx, url, withExperimentalHeader(headers))
}

func (c *experimentalClient) Delete(ctx context.Context, url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	return c.Client.Delete(ctx, url, params, withExperimentalHeader(headers))
}

func (c *experimentalClient) Post(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return c.Client.Post(ctx, url, contentType, body, withExperimentalHeader(headers))
}

func (c *experimentalClient) Put(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return c.Client.Put(ctx, url, contentType, body, withExperimentalHeader(headers))
}

func (c *experimentalClient) Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return c.Client.Patch(ctx, url, body, withExperimentalHeader(headers))
}
