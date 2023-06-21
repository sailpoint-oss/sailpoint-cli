// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Perform Search operations in IdentityNow with a specific query or a template",
		Long:    "\nPerform Search operations in IdentityNow with a specific query or a template\n\n",
		Example: "sail search",
		Aliases: []string{"se"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newQueryCmd(),
		newTemplateCmd(),
	)

	return cmd

}
