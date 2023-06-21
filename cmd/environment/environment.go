package environment

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
)

func NewEnvironmentCmd() *cobra.Command {
	var env string
	var overwrite bool
	var erase bool
	var show bool
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage Environments for the CLI",
		Long:    "\nManage Environments for the CLI\n\n",
		Example: "sail env dev",
		Aliases: []string{"env"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			environments := config.GetEnvironments()
			envKeys := maps.Keys(environments)

			if len(args) > 0 {
				env = args[0]
			} else {
				var choices []tui.Choice
				for i := 0; i < len(envKeys); i++ {
					choices = append(choices, tui.Choice{Title: envKeys[i]})
				}
				selectedEnv, err := tui.PromptList(choices, "Please select an existing environment: ")
				if err != nil {
					return err
				}
				env = selectedEnv.Title
			}

			if env != "" {
				config.SetActiveEnvironment(env)

				if foundEnv, exists := environments[env]; exists && !overwrite && config.GetTenantUrl() != "" && config.GetBaseUrl() != "" {
					if show {
						log.Warn("You are about to Print out the Environment", "env", env)
						res := terminal.InputPrompt("Press Enter to continue")
						if res == "" {
							util.PrettyPrint(foundEnv)
						}
					} else if erase {
						log.Warn("You are about to Erase the Environment", "env", env)
						res := terminal.InputPrompt("Press Enter to continue")
						if res == "" {
							viper.Set("environments."+config.GetActiveEnvironment(), config.Environment{})
						}

					} else {
						log.Info("Environment changed", "env", env)
					}

				} else {

					tenantUrl := terminal.InputPrompt("Tenant URL (ex. https://tenant.identitynow.com):")
					config.SetTenantUrl(tenantUrl)

					baseUrl := terminal.InputPrompt("API Base URL (ex. https://tenant.api.identitynow.com):")
					config.SetBaseUrl(baseUrl)

				}
			} else {
				log.Warn("No Environment Provided")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "use to overwrite an existing environments configuration")
	cmd.Flags().BoolVarP(&erase, "erase", "e", false, "use to erase an existing environments configuration")
	cmd.Flags().BoolVarP(&show, "show", "s", false, "use to show an existing environments configuration")
	cmd.MarkFlagsMutuallyExclusive("overwrite", "erase", "show")

	return cmd

}
