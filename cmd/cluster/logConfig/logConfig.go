// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package logConfig

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed logConfig.md
var logConfigHelp string

func NewLogCommand() *cobra.Command {
	help := util.ParseHelp(logConfigHelp)
	cmd := &cobra.Command{
		Use:     "log",
		Short:   "Interact with a SailPoint Virtual Appliances log configuration",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"l"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newGetCommand(),
		newSetCommand(),
	)

	return cmd
}
