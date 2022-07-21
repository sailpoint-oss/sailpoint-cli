// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnInvokeEntitlementReadCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entitlement-read [id/lookupId] [uniqueId]",
		Short:   "Invoke a std:entitlement:read command",
		Example: `sp connectors invoke entitlement-read john.doe --type group`,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			t := cmd.Flags().Lookup("type").Value.String()

			uniqueID := ""
			if len(args) > 1 {
				uniqueID = args[1]
			}

			_, rawResponse, err := cc.EntitlementRead(ctx, args[0], uniqueID, t)
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
