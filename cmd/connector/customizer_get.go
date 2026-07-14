// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newCustomizerGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get connector customizer",
		Example: "sail conn customizers get -c 1234",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := cmd.Flags().Lookup("id").Value.String()

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			cus, resp, err := apiClient.ConnectorCustomizersAPI.GetConnectorCustomizerV1(cmd.Context(), id).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.Header(toAny(customizerColumns)...)
			table.Append(customizerRow(cus.GetId(), cus.GetName(), cus.ImageVersion))
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Connector customizer ID")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
