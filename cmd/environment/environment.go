package environment

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

func NewEnvironmentCommand() *cobra.Command {
	var env string
	var overwrite bool
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "change currently active environment",
		Long:    "Change Configured Environment that is selected.",
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

				if _, exists := environments[env]; exists && !overwrite && config.GetTenantUrl() != "" && config.GetBaseUrl() != "" {

					log.Log.Info("Environment changed", "env", env)

				} else {

					tenantUrl := terminal.InputPrompt("Tenant URL (ex. https://tenant.identitynow.com):")
					config.SetTenantUrl(tenantUrl)

					baseUrl := terminal.InputPrompt("API Base URL (ex. https://tenant.api.identitynow.com):")
					config.SetBaseUrl(baseUrl)

				}
			} else {
				log.Log.Warn("No Environment Provided")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "use to overwrite an environments configuration")

	return cmd

}
