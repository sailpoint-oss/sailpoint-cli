// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/transform"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all Transforms in IdentityNow",
		Long:    "\nList all Transforms in IdentityNow\n\n",
		Example: "sail transform list | sail transform ls",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			transforms, err := transform.GetTransforms(*apiClient)
			if err != nil {
				return err
			}

			err = transform.ListTransforms(transforms)
			if err != nil {
				return err
			}
			return nil
		},
	}
}
