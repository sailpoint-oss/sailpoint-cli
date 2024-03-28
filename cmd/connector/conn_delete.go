// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

func newConnDeleteCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Connector",
		Long:  "Delete Connector",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			connectorRef := cmd.Flags().Lookup("id").Value.String()
			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()

			resp, err := client.Delete(cmd.Context(), util.ResourceUrl(endpoint, connectorRef))
			if err != nil {
				return err
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("delete connector failed. %s\nbody: %s", resp.Status, body)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "connector %s deleted.\n", connectorRef)
			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Connector ID or Alias")
	_ = cmd.MarkFlagRequired("id")

	bindDevConfig(cmd.Flags())

	return cmd
}
