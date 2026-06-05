package env

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete an environment",
		Long:  "\nDelete a CLI environment and its stored credentials.\nDefaults to the active environment if no name is provided.\n\n",
		Example: `  sail env delete staging
  sail env delete
  sail env delete production --force`,
		Aliases:           []string{"d", "rm"},
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeEnvironmentNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			envName := config.GetActiveEnvironment()
			if len(args) > 0 {
				envName = args[0]
			}

			if envName == "" {
				log.Warn("No active environment configured")
				return clierror.Usage("no active environment configured")
			}

			if environments[envName] == nil {
				log.Warn("Environment does not exist", "name", envName)
				return clierror.NotFound("environment", envName, "Run 'sail env list' to see configured environments.")
			}

			// Safety check: warn if deleting active env
			isActive := envName == config.GetActiveEnvironment()

			if !force {
				msg := fmt.Sprintf("Delete environment '%s'?", envName)
				if isActive {
					msg = fmt.Sprintf("Delete ACTIVE environment '%s'?", envName)
				}
				confirmed, err := tui.Confirm(msg)
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
					return clierror.Canceled("environment delete")
				}
			}

			// Remove from config
			delete(environments, envName)
			viper.Set("environments", environments)

			// Clean up all keyring entries
			auth.DeleteAllPatSecrets(envName)
			auth.DeleteAllOAuthSecrets(envName)

			// If we deleted the active env, switch to another or clear
			if isActive {
				if len(environments) == 0 {
					config.SetActiveEnvironment("")
					log.Info("Environment deleted. No environments remaining.", "deleted", envName)
				} else {
					// Pick the first available environment
					for k := range environments {
						config.SetActiveEnvironment(k)
						log.Info("Environment deleted. Switched active environment.", "deleted", envName, "active", k)
						break
					}
				}
			} else {
				log.Info("Environment deleted", "name", envName)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
