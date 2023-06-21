// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"embed"
	"os"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

//go:embed golang/*
var goTemplateContents embed.FS

const goTemplateDirName = "golang"

func newGolangCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "golang",
		Short:   "Perform Search operations in IdentityNow using a predefined search template",
		Long:    "\nPerform Search operations in IdentityNow using a predefined search template\n\n",
		Example: "sail search template",
		Aliases: []string{"go"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projName string
			var err error

			if len(args) > 0 {
				projName = args[0]
			} else {
				projName, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			err = initialize.InitializeProject(goTemplateContents, goTemplateDirName, projName)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
