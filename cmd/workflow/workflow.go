// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed workflow.md
var workflowHelp string

func NewWorkflowCommand() *cobra.Command {
	help := util.ParseHelp(workflowHelp)
	cmd := &cobra.Command{
		Use:     "workflow",
		Short:   "Manage workflows in Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"work"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(),
		newDownloadCommand(),
		newCreateCommand(),
		newUpdateCommand(),
		newDeleteCommand(),
		newGetCommand(),
	)

	return cmd
}
