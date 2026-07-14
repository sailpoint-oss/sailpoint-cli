// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/golang-sdk/v3/connector_customizers"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newCustomizerUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update connector customizer",
		Example: "sail conn customizers update -c 1234 -n \"My Customizer\"",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := cmd.Flags().Lookup("id").Value.String()
			name := cmd.Flags().Lookup("name").Value.String()

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			request := connector_customizers.Connectorcustomizerupdaterequest{Name: &name}

			cus, resp, err := apiClient.ConnectorCustomizersAPI.PutConnectorCustomizerV1(cmd.Context(), id).Connectorcustomizerupdaterequest(request).Execute()
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

	cmd.Flags().StringP("name", "n", "", "name of the connector customizer")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
