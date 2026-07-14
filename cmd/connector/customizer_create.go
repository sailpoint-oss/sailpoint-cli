// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/golang-sdk/v3/connector_customizers"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newCustomizerCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <customizer-name>",
		Short:   "Create connector customizer",
		Example: "sail conn customizers create \"My Customizer\"",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			request := connector_customizers.Connectorcustomizercreaterequest{Name: &name}

			cus, resp, err := apiClient.ConnectorCustomizersAPI.CreateConnectorCustomizerV1(cmd.Context()).Connectorcustomizercreaterequest(request).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.Header(toAny(customizerColumns)...)
			table.Append(customizerRow(cus.GetId(), cus.GetName(), nil))
			table.Render()

			return nil
		},
	}

	return cmd
}
