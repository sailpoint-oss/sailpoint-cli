package auth

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/spf13/cobra"
)

type authStatus struct {
	Environment   string `json:"environment" yaml:"environment"`
	AuthType      string `json:"authType" yaml:"authType"`
	BaseURL       string `json:"baseUrl" yaml:"baseUrl"`
	TenantURL     string `json:"tenantUrl" yaml:"tenantUrl"`
	TokenStatus   string `json:"tokenStatus" yaml:"tokenStatus"`
	TokenExpiry   string `json:"tokenExpiry,omitempty" yaml:"tokenExpiry,omitempty"`
	RefreshStatus string `json:"refreshStatus,omitempty" yaml:"refreshStatus,omitempty"`
	RefreshExpiry string `json:"refreshExpiry,omitempty" yaml:"refreshExpiry,omitempty"`
	Identity      string `json:"identity,omitempty" yaml:"identity,omitempty"`
	Organization  string `json:"organization,omitempty" yaml:"organization,omitempty"`
}

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
			status := authStatus{
				Environment: envName,
				AuthType:    authType,
				BaseURL:     baseURL,
				TenantURL:   tenantURL,
			}

			switch authType {
			case "pat":
				status = getPATStatus(status, envName)
			case "oauth":
				status = getOAuthStatus(status, envName)
			default:
				status.TokenStatus = "unknown auth type"
			}

			if output.IsMachineReadable() {
				return output.WriteStructured(w, status)
			}

			fmt.Fprintf(w, "Environment:  %s\n", envName)
			fmt.Fprintf(w, "Auth Type:    %s\n", authType)
			fmt.Fprintf(w, "Base URL:     %s\n", baseURL)
			fmt.Fprintf(w, "Tenant URL:   %s\n", tenantURL)

			if status.TokenStatus != "" {
				if status.TokenExpiry != "" {
					fmt.Fprintf(w, "Token:        %s (%s)\n", status.TokenStatus, status.TokenExpiry)
				} else {
					fmt.Fprintf(w, "Token:        %s\n", status.TokenStatus)
				}
			}
			if status.RefreshStatus != "" {
				fmt.Fprintf(w, "Refresh:      %s (%s)\n", status.RefreshStatus, status.RefreshExpiry)
			}
			if status.Identity != "" {
				fmt.Fprintf(w, "Identity:     %v\n", status.Identity)
			}
			if status.Organization != "" {
				fmt.Fprintf(w, "Organization: %v\n", status.Organization)
			}

			return nil
		},
	}
}

func getPATStatus(status authStatus, env string) authStatus {
	expiry, err := auth.GetPatTokenExpiry(env)
	if err != nil {
		status.TokenStatus = "not cached (will authenticate on next command)"
		return status
	}

	status.TokenExpiry = expiry.Format(time.RFC3339)
	if expiry.After(time.Now()) {
		status.TokenStatus = "valid"
	} else {
		status.TokenStatus = "expired"
	}

	token, err := auth.GetPatToken(env)
	if err != nil || token == "" {
		return status
	}

	claims, err := auth.GetTokenClaims(token)
	if err != nil {
		log.Debug("Could not parse token claims", "error", err)
		return status
	}

	if claims["user_name"] != nil {
		status.Identity = fmt.Sprint(claims["user_name"])
	}
	if claims["org"] != nil {
		status.Organization = fmt.Sprint(claims["org"])
	}
	return status
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

func getOAuthStatus(status authStatus, env string) authStatus {
	expiry, err := auth.GetOAuthTokenExpiry(env)
	if err != nil {
		status.TokenStatus = "not cached (will authenticate on next command)"
		return status
	}

	status.TokenExpiry = expiry.Format(time.RFC3339)
	if expiry.After(time.Now()) {
		status.TokenStatus = "valid"
	} else {
		status.TokenStatus = "expired"
	}

	refreshExpiry, err := auth.GetOAuthRefreshExpiry(env)
	if err == nil {
		status.RefreshExpiry = refreshExpiry.Format(time.RFC3339)
		if refreshExpiry.After(time.Now()) {
			status.RefreshStatus = "valid"
		} else {
			status.RefreshStatus = "expired"
		}
	}

	token, err := auth.GetOAuthToken(env)
	if err != nil || token == "" {
		return status
	}

	claims, err := auth.GetTokenClaims(token)
	if err != nil {
		log.Debug("Could not parse token claims", "error", err)
		return status
	}

	if claims["user_name"] != nil {
		status.Identity = fmt.Sprint(claims["user_name"])
	}
	if claims["org"] != nil {
		status.Organization = fmt.Sprint(claims["org"])
	}
	return status
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
