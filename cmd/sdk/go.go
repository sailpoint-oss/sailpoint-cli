// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

const (
	goTemplateRepoOwner = "sailpoint-oss"
	goTemplateRepoName  = "golang-sdk-template"
)

func newGolangCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "golang",
		Short:   "Initialize a new GO SDK project",
		Long:    "\nInitialize a new GO SDK project by fetching the template from GitHub.\n\n",
		Example: "sail sdk init golang\nsail sdk init go example-project",
		Aliases: []string{"go"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projName := "go-template"
			if len(args) > 0 {
				projName = args[0]
			}
			err := initialize.FetchAndInitProject(goTemplateRepoOwner, goTemplateRepoName, "", projName)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Run `cd %s && go mod tidy` to download dependencies.\n", projName)
			return nil
		},
	}
	return cmd
}
