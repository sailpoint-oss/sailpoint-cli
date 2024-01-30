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
					const maxAttempts = 3
					tenant := terminal.InputPrompt("Tenant Name (ie: https://{tenant}.identitynow.com): (" + env + ")")

					if tenant == "" {
						tenant = env
					}

					domain := terminal.InputPrompt("Domain Name: (identitynow.com)")
					tenantUrl := "https://" + tenant + ".identitynow.com"
					baseUrl := "https://" + tenant + ".api.identitynow.com"
					if domain != "" {
						tenantUrl = "https://" + tenant + "." + domain
						baseUrl = "https://" + tenant + ".api." + domain
					}

					authType := terminal.InputPrompt("Authentication Type (oauth, pat):")

					if authType == "pat" {

						clientID, err := config.PromptForClientID()
						if err != nil {
							return err
						}

						ClientSecret, err := config.PromptForClientSecret()
						if err != nil {
							return err
						}

						err = config.SetPatClientSecret(ClientSecret)
						if err != nil {
							return err
						}

						err = config.ResetCachePAT()
						if err != nil {
							return err
						}

						config.SetTenantUrl(tenantUrl)
						config.SetBaseUrl(baseUrl)
						config.SetAuthType(authType)
						config.SetPatClientID(clientID)
					}

					if authType == "oauth" {
						config.SetTenantUrl(tenantUrl)
						config.SetBaseUrl(baseUrl)
						config.SetAuthType(authType)
						config.GetAuthToken()
					}

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
