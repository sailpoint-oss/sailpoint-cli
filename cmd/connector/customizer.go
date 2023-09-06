// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnCustomizersCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customizers",
		Short: "Manage connector customizers",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newCustomizerListCmd(client),
		newCustomizerCreateCmd(client),
		newCustomizerGetCmd(client),
		newCustomizerUpdateCmd(client),
		newCustomizerDeleteCmd(client),
	)

	return cmd
}
