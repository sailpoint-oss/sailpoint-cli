// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

type Config struct {
	ClientID     string
	ClientSecret string
	BaseUrl      string
}

func (c Config) printEnv() {
	fmt.Println("BASE_URL=" + c.BaseUrl)
	fmt.Println("CLIENT_ID=" + c.ClientID)
	fmt.Println("CLIENT_SECRT=" + c.ClientSecret)
}

func newConfigCommand() *cobra.Command {
	var env bool
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Initialize a configuration json file for an SDK project",
		Long:    "\nInitialize a configuration json file for an SDK project\n\nRunning with no arguments will use the currently active environment\n",
		Example: "sail sdk init config\nsail sdk init config <environment name>",
		Aliases: []string{"conf"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			var envName string
			if len(args) > 0 {
				envName = args[0]
			} else {
				envName = config.GetActiveEnvironment()
			}

			clientID, err := config.GetClientID(envName)
			if err != nil {
				return err
			}

			clientSecret, err := config.GetClientSecret(envName)
			if err != nil {
				return err
			}

			SDKConfig := Config{ClientID: clientID, ClientSecret: clientSecret, BaseUrl: config.GetEnvBaseUrl(envName)}

			if env {
				SDKConfig.printEnv()
			} else {
				workingDir, err := os.Getwd()
				if err != nil {
					return err
				}

				configPath := path.Join(workingDir, "config.json")

				file, err := os.Create(configPath)
				if err != nil {
					return err
				}

				defer file.Close()

				configJson, err := json.MarshalIndent(SDKConfig, "", "	")
				if err != nil {
					return err
				}

				_, err = file.Write(configJson)
				if err != nil {
					return err
				}

				log.Info("config file created", "path", configPath)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&env, "environment", "e", false, "Print out the config values in .env format to the terminal rather than to a config file")

	return cmd
}
