// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

const (
	pyTemplateRepoOwner = "sailpoint-oss"
	pyTemplateRepoName  = "python-sdk-template"
)

func newPythonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "python",
		Short:   "Initialize a new python SDK project",
		Long:    "\nInitialize a new Python SDK project by fetching the template from GitHub.\n\n",
		Example: "sail sdk init python\nsail sdk init py example-project",
		Aliases: []string{"py"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projName := "python-template"
			if len(args) > 0 {
				projName = args[0]
			}
			err := initialize.FetchAndInitProject(pyTemplateRepoOwner, pyTemplateRepoName, "", projName)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Run `cd %s && pip install -r requirements.txt` to install dependencies.\n", projName)
			return nil
		},
	}
	return cmd
}
