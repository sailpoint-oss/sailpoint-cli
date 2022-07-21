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

func newConnTagListCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tags for a connector",
		Example: "sp conn tags list -c 1234",
		RunE: func(cmd *cobra.Command, args []string) error {

			connectorRef := cmd.Flags().Lookup("id").Value.String()
			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()

			resp, err := client.Get(cmd.Context(), connResourceUrl(endpoint, connectorRef, "tags"))
			if err != nil {
				return err
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("non-200 response: %s\nbody: %s", resp.Status, body)
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var tags []tag
			err = json.Unmarshal(raw, &tags)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(tagColumns)
			for _, t := range tags {
				table.Append(t.columns())
			}
			table.Render()

			return nil
		},
	}

	return cmd
}
