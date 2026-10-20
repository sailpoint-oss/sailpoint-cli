// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package rule

import (
	"github.com/spf13/cobra"
)

func NewRuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rule",
		Short:   "Manage rules in SailPoint Identity Security",
		Long:    "\nManage rules in SailPoint Identity Security\n\n",
		Example: "sail rule",
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
