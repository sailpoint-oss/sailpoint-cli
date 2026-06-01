package auth

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear cached authentication tokens",
		Long:  "\nClear all cached authentication tokens for the active environment.\nYou will need to re-authenticate on the next API call.\n\n",
		Example: `  sail auth logout
  sail auth logout --env staging`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			envName := config.GetActiveEnvironment()
			if envName == "" {
				return fmt.Errorf("no active environment configured")
			}

			authType := config.GetAuthType()

			if err := auth.Logout(authType, envName); err != nil {
				return fmt.Errorf("failed to clear tokens: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Logged out of '%s'. Cached tokens cleared.\n", envName)
			return nil
		},
	}
}
