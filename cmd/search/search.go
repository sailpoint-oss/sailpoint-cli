// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"github.com/spf13/cobra"
)

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Perform search operations using a query or template",
		Long:    "\nPerform search operations in Identity Security Cloud.\n\nUse 'query' for ad-hoc searches with custom query strings,\nor 'template' to run predefined search templates.\n",
		Example: "  sail search query \"name:a*\" --indices identities\n  sail search template",
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
