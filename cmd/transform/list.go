// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"

	sailpoint "github.com/sailpoint-oss/golang-sdk/v3"
	"github.com/sailpoint-oss/golang-sdk/v3/transforms"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all transforms in Identity Security Cloud",
		Long:    "\nList all transforms in Identity Security Cloud\n\n",
		Example: "sail transform list | sail transform ls",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			transformList, resp, err := sailpoint.PaginateWithDefaults[transforms.Transformread](apiClient.TransformsAPI.ListTransformsV1(context.TODO()))
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			var entries [][]string

			for _, v := range transformList {
				entries = append(entries, []string{v.Name, v.Id})
			}

			output.WriteTable(cmd.OutOrStdout(), []string{"Name", "ID"}, entries, "Name")

			return nil
		},
	}
}
