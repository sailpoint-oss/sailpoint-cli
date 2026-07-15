// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newCustomizerListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all customizers",
		Example: "sail conn customizers list",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			customizers, resp, err := apiClient.ConnectorCustomizersAPI.ListConnectorCustomizersV1(cmd.Context()).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.Header(toAny(customizerColumns)...)
			for _, c := range customizers {
				table.Append(customizerRow(c.GetId(), c.GetName(), c.ImageVersion))
			}
			table.Render()

			return nil
		},
	}

	return cmd
}
