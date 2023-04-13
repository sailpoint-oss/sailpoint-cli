// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSPConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "spconfig",
		Short:   "Perform SPConfig operations in IdentityNow",
		Long:    "\nPerform SPConfig operations in IdentityNow\n\n",
		Example: "sail spconfig",
		Aliases: []string{"spcon"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newExportCmd(),
		newStatusCmd(),
		newTemplateCmd(),
		newDownloadCmd(),
		newImportCommand(),
	)

	return cmd

}
