// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package connector

import (
	"bytes"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

// Unit tests for conn.go

// Expected number of subcommands to `connectors`
const numConnSubcommands = 14

func TestConnResourceUrl(t *testing.T) {
	testEndpoint := "http://localhost:7100/resources"
	testResource := "123"

	expected := "http://localhost:7100/resources/123"
	actual := util.ResourceUrl(testEndpoint, testResource)

	if expected != actual {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestNewConnCmd_noArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := NewConnCmd(mocks.NewMockTerm(ctrl))
	if len(cmd.Commands()) != numConnSubcommands {
		t.Fatalf("expected: %d, actual: %d", len(cmd.Commands()), numConnSubcommands)
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

	if string(out) != cmd.UsageString() {
		t.Errorf("expected: %s, actual: %s", cmd.UsageString(), string(out))
	}
}
