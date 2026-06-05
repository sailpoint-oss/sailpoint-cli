// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

type Config struct {
	ClientId     string
	ClientSecret string
	BaseURL      string
}

func (c Config) printEnv(w io.Writer) {
	fmt.Fprintln(w, "BASE_URL="+c.BaseURL)
	fmt.Fprintln(w, "CLIENT_ID="+c.ClientId)
	fmt.Fprintln(w, "CLIENT_SECRET="+c.ClientSecret)
}

func newConfigCommand() *cobra.Command {
	var env bool
	var unsafePrintSecret bool
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Initialize a configuration JSON file for an SDK project",
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

			SDKConfig := Config{ClientId: clientID, ClientSecret: clientSecret, BaseURL: config.GetEnvBaseUrl(envName)}

			if env {
				if !unsafePrintSecret {
					return fmt.Errorf("--environment prints CLIENT_SECRET and requires --unsafe-print-secret")
				}
				log.Warn("Printing SDK config includes CLIENT_SECRET. Do not paste this output into logs or tickets.")
				SDKConfig.printEnv(cmd.OutOrStdout())
			} else {
				workingDir, err := os.Getwd()
				if err != nil {
					return err
				}

				configPath := path.Join(workingDir, "config.json")

				file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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
	cmd.Flags().BoolVar(&unsafePrintSecret, "unsafe-print-secret", false, "Allow printing CLIENT_SECRET to stdout")

	return cmd
}
