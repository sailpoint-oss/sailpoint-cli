// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newInstanceListCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all connector instances",
		Example: "sail conn instances list",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := client.Get(cmd.Context(), util.ResourceUrl(connectorInstancesEndpoint), nil)
			if err != nil {
				return err
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("list connector instances failed. status: %s\nbody: %s", resp.Status, string(body))
			}

			var instances []instance
			err = json.NewDecoder(resp.Body).Decode(&instances)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(instanceColumns)
			for _, c := range instances {
				table.Append(c.columns())
			}
			table.Render()

			return nil
		},
	}

	return cmd
}
