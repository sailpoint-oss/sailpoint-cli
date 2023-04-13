// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package set

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Configure settings for the SailPoint CLI",
		Long:    "\nConfigure settings for the SailPoint CLI\n\n",
		Example: "sail set",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newDebugCommand(),
		newAuthCommand(),
		newExportTemplateCommand(),
		newSearchTemplateCommand(),
	)

	return cmd

}
