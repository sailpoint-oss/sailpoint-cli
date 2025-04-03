// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package root

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
)

// Expected number of subcommands to `sail` root command
const (
	numRootSubcommands = 15
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
