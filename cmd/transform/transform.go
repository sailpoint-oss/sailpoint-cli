// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"github.com/spf13/cobra"
)

const (
	transformsEndpoint      = "/v3/transforms"
	identityProfileEndpoint = "/v3/identity-profiles"
)

func NewTransformCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transform",
		Short:   "Manage transforms",
		Long:    "\nManage transforms in Identity Security Cloud.\n\nTransforms allow you to manipulate attribute data during provisioning\nand aggregation. Use subcommands to list, create, update, delete,\ndownload, or preview transforms.\n",
		Example: "  sail transform\n  sail tran",
		Aliases: []string{"tran"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.PersistentFlags().StringP("transforms-endpoint", "e", transformsEndpoint, "Override transforms endpoint")

	cmd.AddCommand(
		newListCommand(),
		newDownloadCommand(),
		newCreateCommand(),
		newUpdateCommand(),
		newDeleteCommand(),
		newPreviewCommand(),
	)

	return cmd
}
