// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"github.com/spf13/cobra"
)

func NewSearchCommand() *cobra.Command {
	var folderPath string
	var save bool
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Perform Search operations in IdentityNow with a specific query or a template",
		Long:    "\nPerform Search operations in IdentityNow with a specific query or a template\n\n",
		Example: "sail search",
		Aliases: []string{"se"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newQueryCmd(folderPath, save),
		newTemplateCmd(folderPath, save),
	)

	cmd.PersistentFlags().StringVarP(&folderPath, "folderPath", "f", "", "Folder path to save the search results to. If the directory doesn't exist, then it will be created. (defaults to the current working directory)")
	cmd.PersistentFlags().BoolVarP(&save, "save", "s", false, "Save the search results to a file")

	return cmd

}
