// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package configure

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/spf13/cobra"
)

func NewConfigureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure",
		Short:   "Configure PAT Authentication for the currently active environment",
		Long:    "\nConfigure PAT Authentication for the CLI\n\nPrerequisites:\n\nCreate a Client ID\nCreate a Client Secret\n",
		Aliases: []string{"conf"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			ClientID := terminal.InputPrompt("Personal Access Token Client ID:")
			config.SetPatClientID(ClientID)

			ClientSecret := terminal.InputPrompt("Personal Access Token Client Secret:")
			config.SetPatClientSecret(ClientSecret)

			return nil
		},
	}

	return cmd
}
