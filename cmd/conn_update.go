// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnUpdateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update Connector",
		Long:  "Update Connector",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := cmd.Flags().Lookup("id").Value.String()

			alias := cmd.Flags().Lookup("alias").Value.String()
			if alias == "" {
				return fmt.Errorf("alias must be specified")
			}

			u := connectorUpdate{
				DisplayName: alias,
				Alias:       alias,
			}

			raw, err := json.Marshal(u)
			if err != nil {
				return err
			}

			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()
			resp, err := client.Put(cmd.Context(), connResourceUrl(endpoint, id), "application/json", bytes.NewReader(raw))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("update connector failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err = io.ReadAll(resp.Body)
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

	cmd.Flags().StringP("id", "c", "", "Specify connector id")
	_ = cmd.MarkFlagRequired("id")

	cmd.Flags().StringP("alias", "a", "", "alias of the connector")

	bindDevConfig(cmd.Flags())

	return cmd
}
