// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/sailpoint-oss/sp-cli/client"
	"github.com/sailpoint-oss/sp-cli/cmd/connector"
	"github.com/sailpoint-oss/sp-cli/cmd/transform"
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
		newConfigureCmd(client),
		connector.NewConnCmd(client),
		transform.NewTransformCmd(client),
	)
	return root
}
