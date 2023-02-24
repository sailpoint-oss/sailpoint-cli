// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package set

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "configure settings for the sailpoint cli",
		Long:    "configure settings for the sailpoint cli",
		Example: "sail set",
		Aliases: []string{"set"},
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
