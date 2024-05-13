// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package environment

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "delete",
		Short:   "Delete the active environment in the CLI",
		Long:    "\nDelete the active environment in the CLI\n\n",
		Example: "sail env delete | sail env delete environment_name | sail env d environment_name",
		Aliases: []string{"d"},
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			for _, environmentName := range args {

				if environments[environmentName] != nil {
					log.Warn("You are about to Delete the Environment", "env", environmentName)
					res := terminal.InputPrompt("Press Enter to continue")
					if res == "" {
						delete(environments, environmentName)
						viper.Set("environments", environments)

						config.DeleteOAuthToken(environmentName)
						config.DeleteOAuthTokenExpiry(environmentName)
						config.DeleteRefreshToken(environmentName)
						config.DeleteRefreshTokenExpiry(environmentName)
						config.DeletePatToken(environmentName)
						config.DeletePatTokenExpiry(environmentName)
						config.DeletePatClientID(environmentName)
						config.DeletePatClientSecret(environmentName)

						if len(environments) == 0 {
							config.SetActiveEnvironment("")
						} else {
							for k := range environments {
								config.SetActiveEnvironment(k)
								break
							}
						}

						log.Info("Environment successfully deleted", "environment", environmentName)
					}
				} else {
					log.Warn("Environment does not exist", "env", environmentName)
					return nil
				}
			}

			if len(args) == 0 {
				env := config.GetActiveEnvironment()

				if env != "" && env != " " {

					log.Warn("You are about to Delete the active Environment", "env", env)
					res := terminal.InputPrompt("Press Enter to continue")
					if res == "" {
						delete(environments, env)
						viper.Set("environments", environments)

						config.DeleteOAuthToken("")
						config.DeleteOAuthTokenExpiry("")
						config.DeleteRefreshToken("")
						config.DeleteRefreshTokenExpiry("")
						config.DeletePatToken("")
						config.DeletePatTokenExpiry("")
						config.DeletePatClientID("")
						config.DeletePatClientSecret("")

						if len(environments) == 0 {
							config.SetActiveEnvironment("")
						} else {
							for k := range environments {
								config.SetActiveEnvironment(k)
								break
							}
						}
					}
				} else {
					log.Warn("No environments configured")
					return nil
				}
			}

			return nil
		},
	}
}
