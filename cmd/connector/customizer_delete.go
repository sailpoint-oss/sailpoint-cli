// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newCustomizerDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete connector customizer",
		Example: "sail conn customizers delete -c 1234",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := cmd.Flags().Lookup("id").Value.String()

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			resp, err := apiClient.ConnectorCustomizersAPI.DeleteConnectorCustomizerV1(cmd.Context(), id).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "connector customizer %s deleted.\n", id)
			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Connector customizer ID")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
