// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnInvokeAccountUpdateCmd(spClient client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-update [id/lookupId] [uniqueId] [--changes <value>]",
		Short:   "Invoke a std:account:update command",
		Example: `sp connectors invoke account-update john.doe --changes '[{"op":"Add","attribute":"groups","value":["Group1","Group2"]},{"op":"Set","attribute":"phone","value":2223334444},{"op":"Remove","attribute":"location"}]'`,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, spClient)
			if err != nil {
				return err
			}

			changesRaw := cmd.Flags().Lookup("changes").Value.String()
			var changes []client.AttributeChange
			if err := json.Unmarshal([]byte(changesRaw), &changes); err != nil {
				return err
			}

			uniqueID := ""
			if len(args) > 1 {
				uniqueID = args[1]
			}

			_, rawResponse, err := cc.AccountUpdate(ctx, args[0], uniqueID, changes)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))

			return nil
		},
	}

	cmd.Flags().String("changes", "[]", "Attribute Changes")

	return cmd
}
