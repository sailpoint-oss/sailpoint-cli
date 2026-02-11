// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/spf13/cobra"
)

const (
	pwshTemplateRepoOwner = "sailpoint-oss"
	pwshTemplateRepoName  = "powershell-sdk-template"
)

func newPowerShellCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "powershell",
		Short:   "Initialize a new PowerShell SDK project",
		Long:    "\nInitialize a new PowerShell SDK project by fetching the template from GitHub.\n\n",
		Example: "sail sdk init powershell\nsail sdk init pwsh example-project",
		Aliases: []string{"pwsh"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projName := "powershell-template"
			if len(args) > 0 {
				projName = args[0]
			}
			err := initialize.FetchAndInitProject(pwshTemplateRepoOwner, pwshTemplateRepoName, "", projName)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project created. Configure your SailPoint CLI environment and run the scripts.\n")
			return nil
		},
	}
	return cmd
}
