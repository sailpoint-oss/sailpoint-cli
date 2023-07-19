// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package client

import (
	"os"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

func TestUrlBuilder(t *testing.T) {
	originalURL := os.Getenv("SAIL_BASE_URL")
	os.Setenv("SAIL_BASE_URL", "https://example.com")
	defer os.Setenv("SAIL_BASE_URL", originalURL)

	spClient := &SpClient{
		cfg:    config.CLIConfig{},
		client: nil,
	}

	url := spClient.getUrl("/url/path")
	if url != "https://example.com/url/path" {
		t.Fatalf("expected url to be: \"https://example.com/url/path\", but got: %s", url)
	}

	url = spClient.getUrl("http://localhost:3000")
	if url != "http://localhost:3000" {
		t.Fatalf("expected url to be: \"/http://localhost:3000\", but got: %s", url)
	}
}
