// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnInvokeAccountDiscoverSchemaCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-discover-schema",
		Short:   "Invoke a std:account:discover-schema command",
		Example: `sp connectors invoke account-discover-schema`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			_, rawResponse, err := cc.AccountDiscoverSchema(ctx)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))
			return nil
		},
	}

	return cmd
}
