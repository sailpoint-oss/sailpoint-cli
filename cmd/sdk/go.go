// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

//go:embed golang/*
var goTemplateContents embed.FS

const goTemplateDirName = "golang"

func newGolangCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "golang",
		Short:   "Initialize a new GO SDK project",
		Long:    "\nInitialize a new GO SDK project\n\n",
		Example: "sail sdk init golang\nsail sdk init go example-project",
		Aliases: []string{"go"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projName string
			var err error

			if len(args) > 0 {
				projName = args[0]
			} else {
				projName = "go-template"
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
