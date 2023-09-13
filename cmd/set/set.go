// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package set

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/spf13/cobra"
)

func NewSetCmd(term terminal.Terminal) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Configure settings for the SailPoint CLI",
		Long:    "\nConfigure settings for the SailPoint CLI\n\n",
		Example: "sail set",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newDebugCommand(),
		newAuthCommand(),
		newExportTemplateCommand(),
		newSearchTemplateCommand(),
		newPATCommand(term),
	)

	return cmd

}
