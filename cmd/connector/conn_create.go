// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newConnCreateCmd(client client.Client) *cobra.Command {

	type create struct {
		Alias string `json:"alias"`
	}

	cmd := &cobra.Command{
		Use:     "create <connector-name>",
		Short:   "Create Connector",
		Long:    "Create Connector",
		Example: "sail connectors create \"My-Connector\"",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			if alias == "" {
				return fmt.Errorf("connector alias cannot be empty")
			}

			raw, err := json.Marshal(create{Alias: alias})
			if err != nil {
				return err
			}

			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()
			resp, err := client.Post(cmd.Context(), util.ResourceUrl(endpoint), "application/json", bytes.NewReader(raw), nil)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("create connector failed. status: %s\nbody: %s", resp.Status, body)
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

	return cmd
}
