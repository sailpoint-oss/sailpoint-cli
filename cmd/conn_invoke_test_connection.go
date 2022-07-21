// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnInvokeTestConnectionCmd(spClient client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-connection",
		Short: "Invoke a std:test-connection command",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cc, err := connClient(cmd, spClient)
			if err != nil {
				return err
			}

			rawResponse, err := cc.TestConnection(ctx)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))

			return nil
		},
	}

	return cmd
}
