package environment

import (
	_ "embed"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed environment.md
var environmentHelp string

func NewEnvironmentCommand() *cobra.Command {
	help := util.ParseHelp(environmentHelp)
	var env string
	var overwrite bool
	var erase bool
	var show bool
	var list bool
	var clear bool
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage Environments for the CLI",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"env"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			if len(args) > 0 {
				env = args[0]
			} else {
				env = config.GetActiveEnvironment()
			}

			if env != "" {
				config.SetActiveEnvironment(env)

				if foundEnv, exists := environments[env]; exists && !overwrite && !list && !erase && config.GetTenantUrl() != "" && config.GetBaseUrl() != "" {
					if show {
						log.Warn("You are about to Print out the Environment", "env", env)
						res := terminal.InputPrompt("Press Enter to continue")
						log.Info("Response", "res", res)
						if res == "" {
							fmt.Println(util.PrettyPrint(foundEnv))
						}
					} else if clear {
						log.Warn("You are about to Clear the Environment", "env", env)
						res := terminal.InputPrompt("Press Enter to continue")
						if res == "" {
							viper.Set("environments."+config.GetActiveEnvironment(), config.Environment{})
							config.DeleteOAuthToken()
							config.DeleteOAuthTokenExpiry()
							config.DeleteRefreshToken()
							config.DeleteRefreshTokenExpiry()
							config.DeletePatToken()
							config.DeletePatTokenExpiry()

						}

					} else {
						log.Info("Environment changed", "env", env)
					}

				} else if environments != nil && list {
					log.Warn("You are about to Print out the list of Environments")
					res := terminal.InputPrompt("Press Enter to continue")
					log.Info("Response", "res", res)
					if res == "" {
						fmt.Println(util.PrettyPrint(environments))
					}
				} else if erase {
					log.Warn("You are about to Erase the Environment", "env", env)
					res := terminal.InputPrompt("Press Enter to continue")
					if res == "" {
						delete(environments, env)
						viper.Set("environments", environments)
						config.DeleteOAuthToken()
						config.DeleteOAuthTokenExpiry()
						config.DeleteRefreshToken()
						config.DeleteRefreshTokenExpiry()
						config.DeletePatToken()
						config.DeletePatTokenExpiry()

					}
				} else {

					tenantUrl := terminal.InputPrompt("Tenant URL (ex. https://tenant.identitynow.com):")
					config.SetTenantUrl(tenantUrl)

					baseUrl := terminal.InputPrompt("API Base URL (ex. https://tenant.api.identitynow.com):")
					config.SetBaseUrl(baseUrl)

				}
			} else {
				cmd.Help()
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "use to overwrite an existing environments configuration")
	cmd.Flags().BoolVarP(&erase, "erase", "e", false, "use to erase an existing environment")
	cmd.Flags().BoolVarP(&show, "show", "s", false, "use to show an existing environments configuration")
	cmd.Flags().BoolVarP(&list, "list", "l", false, "use to show a list of envionments")
	cmd.Flags().BoolVarP(&clear, "clear", "c", false, "use to clear an existing environments configuration")
	cmd.MarkFlagsMutuallyExclusive("overwrite", "erase", "show", "list")

	return cmd

}
