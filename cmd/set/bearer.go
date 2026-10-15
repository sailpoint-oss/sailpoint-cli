// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package set

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newBearerCommand() *cobra.Command {
	var token string
	cmd := &cobra.Command{
		Use:   "bearer",
		Short: "Configure bearer token authentication for the currently active environment",
		Long: "\nConfigure bearer token authentication for the CLI\n\n" +
			"Use this when another system has already minted an Identity Security Cloud access token\n" +
			"at the OAuth token endpoint. Pass that access_token instead of a client ID and client secret.\n" +
			"The CLI sends it as a bearer token on every request and does not call the OAuth token endpoint.\n\n" +
			"The token can also be supplied without storing it, for a deployment or a single command:\n\n" +
			"    export SAIL_ACCESS_TOKEN=\"<access_token>\"\n" +
			"    sail --access-token \"<access_token>\" <command>\n\n" +
			"SAIL_ACCESS_TOKEN and --access-token select bearer authentication for that process.\n" +
			"A stored token cannot be refreshed, because the client ID and secret are not available.\n" +
			"Mint a new token when this one expires.",
		Example: "sail set bearer --token <access_token>",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				entered, err := config.PromptForBearerToken()
				if err != nil {
					return err
				}
				token = entered
			}

			if err := config.SetBearerToken(token); err != nil {
				return err
			}

			config.SetAuthType("bearer")
			log.Info("Bearer token stored for the active environment")
			log.Info("Authentication method set to bearer")

			return nil
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "Minted Identity Security Cloud access token to send as a bearer token")

	return cmd
}
