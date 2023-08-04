// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package va

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/va/list"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/va/logConfig"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed va.md
var vaHelp string

func NewVACommand(term terminal.Terminal) *cobra.Command {
	help := util.ParseHelp(vaHelp)
	cmd := &cobra.Command{
		Use:     "va",
		Short:   "Manage Virtual Appliances in IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newCollectCommand(term),
		// newTroubleshootCommand(),
		newParseCommand(),
		newUpdateCommand(term),
		list.NewListCommand(),
		logConfig.NewLogCommand(),
	)

	return cmd
}
