// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package set

import (
	"bufio"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/spf13/cobra"
)

func newPATCommand(term terminal.Terminal) *cobra.Command {
	var ClientID string
	var ClientSecret string
	var readSecretFromStdin bool
	var err error
	cmd := &cobra.Command{
		Use:   "pat",
		Short: "Configure PAT authentication for the currently active environment",
		Long:  "\nConfigure PAT authentication for the CLI\n\nPrerequisites:\n\nCreate a client ID and client secret\nhttps://developer.sailpoint.com/docs/api/authentication#personal-access-tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			if cmd.Flags().Changed("ClientSecret") {
				log.Warn("Passing secrets as flags can expose them in shell history and process listings. Use --client-secret-stdin or the secure prompt instead.")
			}

			if ClientID == "" {

				ClientID, err = config.PromptForClientID()
				if err != nil {
					return err
				}
			}

			err = config.SetPatClientID(ClientID)
			if err != nil {
				return err
			}

			if readSecretFromStdin {
				scanner := bufio.NewScanner(cmd.InOrStdin())
				if scanner.Scan() {
					ClientSecret = strings.TrimSpace(scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					return err
				}
			}

			if ClientSecret == "" {
				ClientSecret, err = config.PromptForClientSecret()
				if err != nil {
					return err
				}
			}

			err = config.SetPatClientSecret(ClientSecret)
			if err != nil {
				return err
			}

			err = config.ResetCachePAT()
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&ClientID, "client-id", "", "The client ID to use for PAT authentication")
	cmd.Flags().BoolVar(&readSecretFromStdin, "client-secret-stdin", false, "Read the client secret from stdin")
	cmd.Flags().StringVarP(&ClientID, "ClientID", "i", "", "Deprecated: use --client-id")
	cmd.Flags().StringVarP(&ClientSecret, "ClientSecret", "s", "", "Deprecated: use --client-secret-stdin or the secure prompt")
	cmd.Flags().MarkDeprecated("ClientID", "use --client-id")
	cmd.Flags().MarkDeprecated("ClientSecret", "use --client-secret-stdin or the secure prompt")

	return cmd
}
