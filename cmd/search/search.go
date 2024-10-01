// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"github.com/spf13/cobra"
)

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Perform search operations in Identity Security Cloud, using a specific query or a template",
		Long:    "\nPerform search operations in Identity Security Cloud, using a specific query or a template\n\n",
		Example: "sail search",
		Aliases: []string{"se"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newQueryCmd(),
		newTemplateCmd(),
	)

	return cmd

}
