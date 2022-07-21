// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnGetCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get Connector",
		Long:  "Get Connector",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			connectorRef := cmd.Flags().Lookup("id").Value.String()

			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()
			resp, err := client.Get(cmd.Context(), connResourceUrl(endpoint, connectorRef))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("get connector failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var conn connector
			err = json.Unmarshal(raw, &conn)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(connectorColumns)
			table.Append(conn.columns())
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Connector ID or Alias")
	_ = cmd.MarkFlagRequired("id")

	bindDevConfig(cmd.Flags())

	return cmd
}
