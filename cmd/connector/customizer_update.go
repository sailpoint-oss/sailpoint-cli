// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
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

func newCustomizerUpdateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Create connector customizer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			id := cmd.Flags().Lookup("id").Value.String()
			name := cmd.Flags().Lookup("name").Value.String()

			raw, err := json.Marshal(customizer{Name: name})
			if err != nil {
				return err
			}

			resp, err := client.Put(cmd.Context(), util.ResourceUrl(connectorCustomizersEndpoint, id), "application/json", bytes.NewReader(raw))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("create customizer failed. status: %s\nbody: %s", resp.Status, string(body))
			}

			var cus customizer
			err = json.NewDecoder(resp.Body).Decode(&cus)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(customizerColumns)
			table.Append(cus.columns())
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Specify connector customizer id")
	_ = cmd.MarkFlagRequired("id")

	cmd.Flags().StringP("name", "n", "", "name of the connector customizer")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
