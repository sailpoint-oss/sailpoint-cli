// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

//go:embed python/*
var pyTemplateContents embed.FS

const pyTemplateDirName = "python"

func newPythonCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "python",
		Short:   "Initialize a new python SDK project",
		Long:    "\nInitialize a new typescript SDK project\n\n",
		Example: "sail sdk init python\nsail sdk init py example-project",
		Aliases: []string{"py"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projName string
			var err error

			if len(args) > 0 {
				projName = args[0]
			} else {
				projName = "python-template"
			}

			err = initialize.InitializeProject(pyTemplateContents, pyTemplateDirName, projName)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
