// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package rule

import (
	"github.com/spf13/cobra"
)

func NewRuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rule",
		Short:   "Manage rules",
		Long:    "\nManage cloud rules and connector rules in Identity Security Cloud.\nUse subcommands to list or download rules.\n",
		Example: "  sail rule list\n  sail rule download",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(),
		newDownloadCommand(),
	)

	return cmd
}
