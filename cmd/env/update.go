package env

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update [name]",
		Short: "Update an existing environment",
		Long:  "\nUpdate the configuration of an existing environment.\nDefaults to the active environment if no name is provided.\n\n",
		Example: `  sail env update
  sail env update production`,
		Aliases: []string{"up"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			envName := config.GetActiveEnvironment()
			if len(args) > 0 {
				envName = args[0]
			}

			if envName == "" {
				log.Warn("No active environment configured. Run 'sail env create' to create one.")
				return nil
			}

			if environments[envName] == nil {
				log.Warn("Environment does not exist", "name", envName,
					"hint", "Use 'sail env create "+envName+"' to create it.")
				return nil
			}

			return createOrUpdateEnv([]string{envName}, true)
		},
	}
}
