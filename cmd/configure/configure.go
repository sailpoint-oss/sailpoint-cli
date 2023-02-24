// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package configure

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConfigureCmd() *cobra.Command {
	var debug bool
	cmd := &cobra.Command{
		Use:     "configure",
		Short:   "configure pat authentication for the currently active environment",
		Long:    "\nConfigure PAT Authentication for the CLI\n\nPrerequisites:\n\nClient ID\nClient Secret\n",
		Aliases: []string{"conf"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			err := config.InitConfig()
			if err != nil {
				return err
			}

			viper.Set("debug", debug)

			ClientID := terminal.InputPrompt("Personal Access Token Client ID:")
			ClientSecret := terminal.InputPrompt("Personal Access Token Client Secret:")

			config.SetPatClientID(ClientID)
			config.SetPatClientSecret(ClientSecret)

			err = config.SaveConfig()
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Specifies if the debug flag should be set")

	return cmd
}
