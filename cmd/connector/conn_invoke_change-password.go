// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnInvokeChangePasswordCmd(spClient client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "change-password",
		Short:   "Invoke a change-password command",
		Example: `sail connectors invoke change-password john.doe newPassword`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cc, err := connClient(cmd, spClient)
			if err != nil {
				return err
			}

			rawResponse, err := cc.ChangePassword(ctx, args[0], "", args[1])
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))

			return nil
		},
	}

	return cmd
}
