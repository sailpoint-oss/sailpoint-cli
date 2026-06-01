package auth

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		Long:  "\nDisplay information about the current authentication session,\nincluding the active environment, auth type, identity, and token expiry.\n\n",
		Example: `  sail auth status
  sail auth status --env production`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			envName := config.GetActiveEnvironment()
			if envName == "" {
				fmt.Fprintln(w, "No active environment configured.")
				return nil
			}

			authType := config.GetAuthType()
			baseURL := config.GetBaseUrl()
			tenantURL := config.GetTenantUrl()

			fmt.Fprintf(w, "Environment:  %s\n", envName)
			fmt.Fprintf(w, "Auth Type:    %s\n", authType)
			fmt.Fprintf(w, "Base URL:     %s\n", baseURL)
			fmt.Fprintf(w, "Tenant URL:   %s\n", tenantURL)

			switch authType {
			case "pat":
				showPATStatus(w, envName)
			case "oauth":
				showOAuthStatus(w, envName)
			default:
				fmt.Fprintf(w, "Auth Type:    %s (unknown)\n", authType)
			}

			return nil
		},
	}
}

func showPATStatus(w io.Writer, env string) {
	expiry, err := auth.GetPatTokenExpiry(env)
	if err != nil {
		fmt.Fprintln(w, "Token:        not cached (will authenticate on next command)")
		return
	}

	if expiry.After(time.Now()) {
		fmt.Fprintf(w, "Token:        valid (expires %s)\n", expiry.Format(time.RFC3339))
	} else {
		fmt.Fprintf(w, "Token:        expired (expired %s)\n", expiry.Format(time.RFC3339))
	}

	token, err := auth.GetPatToken(env)
	if err != nil || token == "" {
		return
	}

	claims, err := auth.GetTokenClaims(token)
	if err != nil {
		log.Debug("Could not parse token claims", "error", err)
		return
	}

	if claims["user_name"] != nil {
		fmt.Fprintf(w, "Identity:     %v\n", claims["user_name"])
	}
	if claims["org"] != nil {
		fmt.Fprintf(w, "Organization: %v\n", claims["org"])
	}
}

func showOAuthStatus(w io.Writer, env string) {
	expiry, err := auth.GetOAuthTokenExpiry(env)
	if err != nil {
		fmt.Fprintln(w, "Token:        not cached (will authenticate on next command)")
		return
	}

	if expiry.After(time.Now()) {
		fmt.Fprintf(w, "Token:        valid (expires %s)\n", expiry.Format(time.RFC3339))
	} else {
		fmt.Fprintf(w, "Token:        expired (expired %s)\n", expiry.Format(time.RFC3339))
	}

	refreshExpiry, err := auth.GetOAuthRefreshExpiry(env)
	if err == nil {
		if refreshExpiry.After(time.Now()) {
			fmt.Fprintf(w, "Refresh:      valid (expires %s)\n", refreshExpiry.Format(time.RFC3339))
		} else {
			fmt.Fprintf(w, "Refresh:      expired (expired %s)\n", refreshExpiry.Format(time.RFC3339))
		}
	}

	token, err := auth.GetOAuthToken(env)
	if err != nil || token == "" {
		return
	}

	claims, err := auth.GetTokenClaims(token)
	if err != nil {
		log.Debug("Could not parse token claims", "error", err)
		return
	}

	if claims["user_name"] != nil {
		fmt.Fprintf(w, "Identity:     %v\n", claims["user_name"])
	}
	if claims["org"] != nil {
		fmt.Fprintf(w, "Organization: %v\n", claims["org"])
	}
}
