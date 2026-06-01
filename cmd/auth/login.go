package auth

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the active environment",
		Long:  "\nExplicitly authenticate with the active environment using its configured\nauthentication method (PAT or OAuth).\n\nFor PAT environments, this uses the stored Client ID and Client Secret.\nFor OAuth environments, this opens a browser for interactive login.\n\n",
		Example: `  sail auth login
  sail auth login --env production`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			envName := config.GetActiveEnvironment()
			if envName == "" {
				return fmt.Errorf("no active environment configured. Run 'sail env create' first")
			}

			authType := config.GetAuthType()
			baseURL := config.GetBaseUrl()
			tenantURL := config.GetTenantUrl()
			tokenURL := config.GetTokenUrl()

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Logging in to '%s' (%s)...\n", envName, authType)

			token, err := auth.GetToken(authType, envName, baseURL, tenantURL, tokenURL, func(newBaseURL string) {
				config.SetBaseUrl(newBaseURL)
			})
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			claims, err := auth.GetTokenClaims(token)
			if err != nil {
				log.Warn("Could not parse token claims", "error", err)
				fmt.Fprintln(w, "Login successful.")
				return nil
			}

			userName := claims["user_name"]
			org := claims["org"]
			fmt.Fprintf(w, "Authenticated as: %v (org: %v)\n", userName, org)

			return nil
		},
	}
}
