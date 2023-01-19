// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnInvokeChangePasswordCmd(spClient client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "change-password",
		Short:   "Invoke a change-password command",
		Example: `sail connectors invoke change-password john.doe`,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cc, err := connClient(cmd, spClient)
			if err != nil {
				return err
			}

			// prompt the user for the password
			prompt := promptui.Prompt{
				Label: "Enter the password",
				Mask:  '*',
			}

			password, err := prompt.Run()
			if err != nil {
				return err
			}

			// uniqueID if provided
			uniqueID := ""
			if len(args) > 1 {
				uniqueID = args[1]
			}

			rawResponse, err := cc.ChangePassword(ctx, args[0], uniqueID, password)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))

			return nil
		},
	}

	return cmd
}
