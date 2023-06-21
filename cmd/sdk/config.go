// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

type Config struct {
	ClientID     string
	ClientSecret string
	BaseUrl      string
	TokenUrl     string
}

func newConfigCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Perform Search operations in IdentityNow using a predefined search template",
		Long:    "\nPerform Search operations in IdentityNow using a predefined search template\n\n",
		Example: "sail search template",
		Aliases: []string{"conf"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			SDKConfig := Config{ClientID: config.GetPatClientID(), ClientSecret: config.GetPatClientSecret()}

			return nil
		},
	}
	return cmd
}
