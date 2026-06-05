package env

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/redact"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newShowCommand() *cobra.Command {
	var showRaw bool
	type environmentDetails struct {
		Name      string `json:"name" yaml:"name"`
		Active    bool   `json:"active" yaml:"active"`
		TenantURL string `json:"tenantUrl" yaml:"tenantUrl"`
		BaseURL   string `json:"baseUrl" yaml:"baseUrl"`
		AuthType  string `json:"authType" yaml:"authType"`
	}

	cmd := &cobra.Command{
		Use:               "show [name]",
		Short:             "Show details of an environment",
		Long:              "\nShow the configuration details of an environment.\nDefaults to the active environment if no name is provided.\n\n",
		Example:           "sail env show\nsail env show production",
		Aliases:           []string{"s"},
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeEnvironmentNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			envName := config.GetActiveEnvironment()
			if len(args) > 0 {
				envName = args[0]
			}

			if envName == "" || (envName != "" && environments[envName] == nil) {
				if envName == "" {
					log.Warn("No active environment configured. Run 'sail env create' to create one.")
				} else {
					log.Warn("Environment does not exist", "name", envName)
				}
				return clierror.NotFound("environment", envName, "Run 'sail env list' to see configured environments.")
			}

			details := environmentDetails{
				Name:      envName,
				Active:    envName == config.GetActiveEnvironment(),
				TenantURL: config.GetEnvTenantUrl(envName),
				BaseURL:   config.GetEnvBaseUrl(envName),
				AuthType:  config.GetEnvAuthType(envName),
			}

			if output.IsMachineReadable() {
				return output.WriteStructured(cmd.OutOrStdout(), details)
			}

			activeIndicator := ""
			if details.Active {
				activeIndicator = " (active)"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Environment: %s%s\n", envName, activeIndicator)
			fmt.Fprintf(cmd.OutOrStdout(), "  Tenant URL: %s\n", details.TenantURL)
			fmt.Fprintf(cmd.OutOrStdout(), "  Base URL:   %s\n", details.BaseURL)
			fmt.Fprintf(cmd.OutOrStdout(), "  Auth Type:  %s\n", details.AuthType)

			if showRaw {
				raw := redact.Value(environments[envName])
				fmt.Fprintf(cmd.OutOrStdout(), "\nRaw config (redacted):\n%s\n", util.PrettyPrint(raw))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showRaw, "raw", false, "Show redacted raw environment configuration")
	return cmd
}
