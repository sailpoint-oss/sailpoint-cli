// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnInvokeAccountCreateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-create [identity] [--attributes <value>]",
		Short:   "Invoke a std:account:create command",
		Example: `sp connectors invoke account-create john.doe --attributes '{"email": "john.doe@example.com"}'`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			var identity *string = nil
			if len(args) > 0 {
				identity = &args[0]
			}

			attributesRaw := cmd.Flags().Lookup("attributes").Value.String()
			var attributes map[string]interface{}
			if err := json.Unmarshal([]byte(attributesRaw), &attributes); err != nil {
				return err
			}

			_, rawResponse, err := cc.AccountCreate(ctx, identity, attributes)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))

			return nil
		},
	}

	cmd.Flags().StringP("attributes", "a", "{}", "Attributes")

	return cmd
}
