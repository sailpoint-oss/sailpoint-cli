// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnInvokeEntitlementListCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entitlement-list [--type <value>]",
		Short:   "Invoke a std:entitlement:list command",
		Example: `sp connectors invoke entitlement-list --type group`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			t := cmd.Flags().Lookup("type").Value.String()
			_, rawResponse, err := cc.EntitlementList(ctx, t)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))

			return nil
		},
	}

	cmd.Flags().StringP("type", "t", "", "Entitlement Type")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}
