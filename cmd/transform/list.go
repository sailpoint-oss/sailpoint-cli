// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"

	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	v3 "github.com/sailpoint-oss/golang-sdk/v2/api_v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all transforms in IdentityNow",
		Long:    "\nList all transforms in IdentityNow\n\n",
		Example: "sail transform list | sail transform ls",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			transforms, resp, err := sailpoint.PaginateWithDefaults[v3.TransformRead](apiClient.V3.TransformsAPI.ListTransforms(context.TODO()))
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			var entries [][]string

			for _, v := range transforms {
				entries = append(entries, []string{v.Name, v.Id})
			}

			output.WriteTable(cmd.OutOrStdout(), []string{"Name", "ID"}, entries)

			return nil
		},
	}
}
