package env

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "use <name>",
		Short:   "Switch the active environment",
		Long:    "\nSet an environment as the active environment for CLI commands.\n\n",
		Example: "sail env use production",
		Aliases: []string{"u"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			envName := args[0]
			environments := config.GetEnvironments()

			if environments[envName] == nil {
				log.Warn("Environment does not exist",
					"name", envName,
					"hint", "Use 'sail env create "+envName+"' to create it.")
				return nil
			}

			config.SetActiveEnvironment(envName)
			authType := config.GetEnvAuthType(envName)
			tenantURL := config.GetEnvTenantUrl(envName)

			log.Info("Switched active environment",
				"env", envName,
				"tenant", tenantURL,
				"auth", authType)

			return nil
		},
	}
}
