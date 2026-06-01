package env

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "show [name]",
		Short:   "Show details of an environment",
		Long:    "\nShow the configuration details of an environment.\nDefaults to the active environment if no name is provided.\n\n",
		Example: "sail env show\nsail env show production",
		Aliases: []string{"s"},
		Args:    cobra.MaximumNArgs(1),
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
				return nil
			}

			activeIndicator := ""
			if envName == config.GetActiveEnvironment() {
				activeIndicator = " (active)"
			}

			fmt.Printf("Environment: %s%s\n", envName, activeIndicator)
			fmt.Printf("  Tenant URL: %s\n", config.GetEnvTenantUrl(envName))
			fmt.Printf("  Base URL:   %s\n", config.GetEnvBaseUrl(envName))
			fmt.Printf("  Auth Type:  %s\n", config.GetEnvAuthType(envName))

			// Show the full raw config for detailed view
			fmt.Printf("\nRaw config:\n%s\n", util.PrettyPrint(environments[envName]))

			return nil
		},
	}
}
