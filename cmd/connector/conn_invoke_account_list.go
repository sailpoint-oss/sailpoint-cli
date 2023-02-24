// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnInvokeAccountListCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-list",
		Short: "Invoke a std:account:list command",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			_, state, printable, err := cc.AccountList(ctx)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Accounts:")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(printable))

			if state != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nState:\n")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(state))
			}
			return nil
		},
	}

	return cmd
}
