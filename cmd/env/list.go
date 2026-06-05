package env

import (
	"sort"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	type environmentRow struct {
		Active    bool   `json:"active" yaml:"active"`
		Name      string `json:"name" yaml:"name"`
		TenantURL string `json:"tenantUrl" yaml:"tenantUrl"`
		AuthType  string `json:"authType" yaml:"authType"`
	}

	return &cobra.Command{
		Use:     "list",
		Short:   "List all configured environments",
		Long:    "\nList all configured environments with their tenant URLs and auth types.\n\n",
		Example: "sail env list",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			if len(environments) == 0 {
				log.Warn("No environments configured. Run 'sail env create' to create one.")
				return nil
			}

			activeEnv := config.GetActiveEnvironment()

			headers := []string{"", "Name", "Tenant URL", "Auth Type"}
			var rows [][]string
			var structuredRows []environmentRow

			// Sort env names for stable output
			names := make([]string, 0, len(environments))
			for name := range environments {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				active := ""
				if name == activeEnv {
					active = "*"
				}
				tenantURL := config.GetEnvTenantUrl(name)
				authType := config.GetEnvAuthType(name)
				if authType == "" {
					authType = "pat"
				}
				rows = append(rows, []string{active, name, tenantURL, authType})
				structuredRows = append(structuredRows, environmentRow{
					Active:    name == activeEnv,
					Name:      name,
					TenantURL: tenantURL,
					AuthType:  authType,
				})
			}

			return output.WriteTableOrStructured(cmd.OutOrStdout(), headers, rows, "", structuredRows)
		},
	}
}
