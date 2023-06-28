// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

//go:embed powershell/*
var pwshTemplateContents embed.FS

const pwshTemplateDirName = "powershell"

func newPowerShellCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "powershell",
		Short:   "Initialize a new PowerShell SDK project",
		Long:    "\nInitialize a new PowerShell SDK project\n\n",
		Example: "sail sdk init powershell\nsail sdk init pwsh example-project",
		Aliases: []string{"pwsh"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projName string
			var err error

			if len(args) > 0 {
				projName = args[0]
			} else {
				projName = "powershell-template"
			}

			err = initialize.InitializeProject(pwshTemplateContents, pwshTemplateDirName, projName)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
