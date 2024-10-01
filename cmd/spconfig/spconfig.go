// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed spconfig.md
var spconfigHelp string

func NewSPConfigCommand() *cobra.Command {
	help := util.ParseHelp(spconfigHelp)
	cmd := &cobra.Command{
		Use:     "spconfig",
		Short:   "Perform SPConfig operations in Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"spcon"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newExportCommand(),
		newStatusCommand(),
		newTemplateCommand(),
		newDownloadCommand(),
		newImportCommand(),
	)

	return cmd

}
