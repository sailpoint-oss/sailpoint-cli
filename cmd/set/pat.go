// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package set

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/spf13/cobra"
)

func newPATCommand(term terminal.Terminal) *cobra.Command {
	var ClientID string
	var ClientSecret string
	var err error
	cmd := &cobra.Command{
		Use:   "pat",
		Short: "Configure PAT authentication for the currently active environment",
		Long:  "\nConfigure PAT authentication for the CLI\n\nPrerequisites:\n\nCreate a client ID and client secret\nhttps://developer.sailpoint.com/idn/api/authentication#personal-access-tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			if ClientID == "" {
				ClientID, err = term.PromptPassword("Personal Access Token Client ID:")
				if err != nil {
					return err
				}
			}

			config.SetPatClientID(ClientID)

			if ClientSecret == "" {
				ClientSecret, err = term.PromptPassword("Personal Access Token Client Secret:")
				if err != nil {
					return err
				}
			}

			config.SetPatClientSecret(ClientSecret)

			return nil
		},
	}

	cmd.Flags().StringVarP(&ClientID, "ClientID", "i", "", "The client id to use for PAT authentication")
	cmd.Flags().StringVarP(&ClientSecret, "ClientSecret", "s", "", "The client secret to use for PAT authentication")

	return cmd
}
