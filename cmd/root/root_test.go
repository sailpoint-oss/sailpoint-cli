// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package root

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestNewRootCmd_noArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := NewRootCommand()
	assertRootCommands(t, cmd, []string{
		"access-profile",
		"access-request",
		"account",
		"api",
		"auth",
		"cluster",
		"config",
		"connectors",
		"entitlement",
		"env",
		"identity",
		"role",
		"source",
	})
	assertNoRootCommands(t, cmd, []string{
		"access",
		"admin",
		"apps",
		"audit",
		"lifecycle",
		"users",
	})

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

func assertRootCommands(t *testing.T, cmd interface{ CommandPath() string }, expected []string) {
	t.Helper()
	rootCmd, ok := cmd.(interface {
		Find([]string) (*cobra.Command, []string, error)
	})
	if !ok {
		t.Fatalf("unexpected command type")
	}
	for _, name := range expected {
		found, _, err := rootCmd.Find([]string{name})
		if err != nil || found == nil || found.Name() != name {
			t.Fatalf("expected root command %q to exist", name)
		}
	}
}

func assertNoRootCommands(t *testing.T, cmd interface{ CommandPath() string }, names []string) {
	t.Helper()
	rootCmd, ok := cmd.(interface {
		Find([]string) (*cobra.Command, []string, error)
	})
	if !ok {
		t.Fatalf("unexpected command type")
	}
	for _, name := range names {
		found, _, err := rootCmd.Find([]string{name})
		if err == nil && found != nil && found.Name() == name {
			t.Fatalf("did not expect root command %q to exist", name)
		}
	}
}

func TestNewRootCmd_completionEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := NewRootCommand()

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"completion", "bash"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected completion command to succeed: %v", err)
	}
}

func TestRootEnvFlagDoesNotPersistActiveEnvironment(t *testing.T) {
	viper.Reset()
	config.ClearActiveEnvironmentOverride()
	t.Cleanup(func() {
		viper.Reset()
		config.ClearActiveEnvironmentOverride()
	})

	viper.Set("activeenvironment", "production")

	cmd := NewRootCommand()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--env", "staging", "config", "debug"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected command to succeed: %v", err)
	}

	if got := config.GetActiveEnvironment(); got != "staging" {
		t.Fatalf("active environment override = %q, want %q", got, "staging")
	}
	if got := viper.GetString("activeenvironment"); got != "production" {
		t.Fatalf("persisted activeenvironment = %q, want %q", got, "production")
	}
}
