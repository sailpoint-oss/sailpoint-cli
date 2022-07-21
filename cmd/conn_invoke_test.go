// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint/sp-cli/mocks"
)

// Unit tests for conn_invoke.go and its subcommands

// Expected number of subcommands to `sp` root command
const numConnInvokeSubcommands = 10

func TestNewConnInvokeCmd_noArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := newConnInvokeCmd(mocks.NewMockClient(ctrl))
	if len(cmd.Commands()) != numConnInvokeSubcommands {
		t.Fatalf("expected: %d, actual: %d", len(cmd.Commands()), numConnInvokeSubcommands)
	}

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{})
	cmd.PersistentFlags().Set("id", "connector-id")
	cmd.PersistentFlags().Set("version", "455455")

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("error execute cmd: %v", err)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("error read out: %v", err)
	}

	if string(out) != cmd.UsageString() {
		t.Errorf("expected: %s, actual: %s", cmd.UsageString(), string(out))
	}
}
