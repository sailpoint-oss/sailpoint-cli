package auth

import (
	"bufio"
	"fmt"
	"strings"

	internalauth "github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newPATCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pat",
		Short: "Manage PAT credentials for the active environment",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(newPATSetCommand())
	return cmd
}

func newPATSetCommand() *cobra.Command {
	var clientID string
	var readSecretFromStdin bool
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Set PAT credentials for the active environment",
		Long:    "\nSet PAT credentials for the active environment. Use --client-secret-stdin for non-interactive setup without placing secrets in shell history.\n\n",
		Example: `  printf '%s' "$SAIL_CLIENT_SECRET" | sail auth pat set --client-id "$SAIL_CLIENT_ID" --client-secret-stdin`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			env := config.GetActiveEnvironment()
			if clientID == "" {
				var err error
				clientID, err = internalauth.PromptForClientID()
				if err != nil {
					return err
				}
			}

			clientSecret := ""
			if readSecretFromStdin {
				scanner := bufio.NewScanner(cmd.InOrStdin())
				if scanner.Scan() {
					clientSecret = strings.TrimSpace(scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					return err
				}
			}
			if clientSecret == "" {
				var err error
				clientSecret, err = internalauth.PromptForClientSecret()
				if err != nil {
					return err
				}
			}

			if err := internalauth.SetPatClientID(env, clientID); err != nil {
				return err
			}
			if err := internalauth.SetPatClientSecret(env, clientSecret); err != nil {
				return err
			}
			if err := internalauth.ResetCachePAT(env); err != nil {
				return err
			}

			_, err := fmt.Fprintf(cmd.ErrOrStderr(), "PAT credentials saved for environment %q\n", env)
			return err
		},
	}
	cmd.Flags().StringVar(&clientID, "client-id", "", "PAT client ID")
	cmd.Flags().BoolVar(&readSecretFromStdin, "client-secret-stdin", false, "Read PAT client secret from stdin")
	return cmd
}
