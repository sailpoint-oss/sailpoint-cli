// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

const (
	tsTemplateRepoOwner = "sailpoint-oss"
	tsTemplateRepoName  = "typescript-sdk-template"
)

func newTypescriptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "typescript",
		Short:   "Initialize a new typescript SDK project",
		Long:    "\nInitialize a new TypeScript SDK project by fetching the template from GitHub.\n\n",
		Example: "sail sdk init typescript\nsail sdk init ts example-project",
		Aliases: []string{"ts"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projName := "typescript-template"
			if len(args) > 0 {
				projName = args[0]
			}
			err := initialize.FetchAndInitProject(tsTemplateRepoOwner, tsTemplateRepoName, "", projName)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Run `cd %s && npm install` to install dependencies.\n", projName)
			return nil
		},
	}
	return cmd
}
