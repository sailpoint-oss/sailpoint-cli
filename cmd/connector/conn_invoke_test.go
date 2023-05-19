// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package connector

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
	"github.com/spf13/cobra"
)

// Unit tests for conn_invoke.go and its subcommands

// Expected number of subcommands to `sp` root command
const numConnInvokeSubcommands = 11

func TestNewConnInvokeCmd_noArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := newConnInvokeCmd(mocks.NewMockClient(ctrl), mocks.NewMockTerm(ctrl))
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

func TestInvokeConfig(t *testing.T) {
	t.Run("Both config-path and config-json are empty", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("config-path", "", "")
		cmd.Flags().String("config-json", "", "")

		_, err := invokeConfig(cmd)
		expectedErr := fmt.Errorf("Either config-path or config-json must be set")
		if err.Error() != expectedErr.Error() {
			t.Fatalf("expected err: %s, actual: %s", expectedErr, err)
		}
	})

	t.Run("Using config-json", func(t *testing.T) {
		configContent := `{"key": "value"}`

		cmd := &cobra.Command{}
		cmd.Flags().String("config-path", "", "")
		cmd.Flags().String("config-json", configContent, "")

		_, err := invokeConfig(cmd)

		if err != nil {
			t.Fatalf("expected nil err, actual: %s", err)
		}
	})

	t.Run("Using config-path", func(t *testing.T) {
		// temporary file to be used as a config-path flag value
		tempFile, err := ioutil.TempFile("", "temp-config-*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tempFile.Name())

		configContent := `{"key": "value"}`
		if _, err := tempFile.Write([]byte(configContent)); err != nil {
			t.Fatal(err)
		}
		if err := tempFile.Close(); err != nil {
			t.Fatal(err)
		}

		// run command
		cmd := &cobra.Command{}
		cmd.Flags().String("config-path", tempFile.Name(), "")
		cmd.Flags().String("config-json", "", "")

		_, err = invokeConfig(cmd)

		if err != nil {
			t.Fatalf("expected nil err, actual: %s", err)
		}
	})
}
