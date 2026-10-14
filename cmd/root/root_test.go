// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package root

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

// Expected number of subcommands to `sail` root command
const (
	numRootSubcommands = 17
)

func TestNewRootCmd_noArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := NewRootCommand()
	if len(cmd.Commands()) != numRootSubcommands {
		t.Fatalf("expected: %d, actual: %d", numRootSubcommands, len(cmd.Commands()))
	}

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("error execute cmd: %v", err)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("error read out: %v", err)
	}

	if !strings.Contains(string(out), cmd.UsageString()) {
		t.Errorf("expected: %s, actual: %s", cmd.UsageString(), string(out))
	}
}

func TestNewRootCmd_accessTokenFlag(t *testing.T) {
	t.Setenv("SAIL_ACCESS_TOKEN", "")
	t.Cleanup(func() { config.SetRuntimeAccessToken("") })

	payload, err := json.Marshal(map[string]any{
		"org": "acme",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	token := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`)) + "." +
		base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString([]byte("sig"))

	cmd := NewRootCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--access-token", "Bearer " + token})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got, err := config.GetBearerToken()
	if err != nil {
		t.Fatalf("GetBearerToken() error = %v", err)
	}
	if got != token {
		t.Fatalf("GetBearerToken() = %q, want the access token passed on the command line", got)
	}
}

func TestNewRootCmd_completionDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := NewRootCommand()

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"completion"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail")
	}
}
