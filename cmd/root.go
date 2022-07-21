// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func NewRootCmd(client client.Client) *cobra.Command {
	root := &cobra.Command{
		Use:          "sp",
		Short:        "sp",
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}
	root.AddCommand(
		newConnCmd(client),
		newConfigureCmd(client),
	)
	return root
}
