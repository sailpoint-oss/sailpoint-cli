// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"fmt"
	"os"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

func TestMain(m *testing.M) {
	if err := config.InitConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize CLI config: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func requireLiveCredentials(t *testing.T) {
	t.Helper()

	if err := config.Validate(); err != nil {
		t.Skipf("skipping live API test: no usable SailPoint CLI credentials found (%v). Configure PAT credentials with SAIL_BASE_URL, SAIL_CLIENT_ID, and SAIL_CLIENT_SECRET, or run `sail env create`/`sail auth login` for OAuth, then rerun this test.", err)
	}
}
