// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"github.com/spf13/cobra"
)

func NewSDKCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sdk",
		Short:   "Initialize or configure SDK projects",
		Long:    "\nInitialize or configure SailPoint SDK projects.\n\nSupported languages: Go, Python, TypeScript, and PowerShell.\n",
		Example: "  sail sdk init\n  sail sdk init config",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newInitCommand(),
	)

	return cmd

}
