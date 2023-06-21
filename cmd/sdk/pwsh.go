// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"embed"
	"os"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

//go:embed powershell/*
var pwshTemplateContents embed.FS

const pwshTemplateDirName = "powershell"

func newPowerShellCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "powershell",
		Short:   "Perform Search operations in IdentityNow using a predefined search template",
		Long:    "\nPerform Search operations in IdentityNow using a predefined search template\n\n",
		Example: "sail search template",
		Aliases: []string{"pwsh"},
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

			err = initialize.InitializeProject(pwshTemplateContents, pwshTemplateDirName, projName)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
