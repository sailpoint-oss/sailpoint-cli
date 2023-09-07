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

func newCustomizerUnlinkCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unlink",
		Short:   "Unlink connector customizer from connector instance",
		Example: "sail conn customizers unlink -i 5678",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceID := cmd.Flags().Lookup("instance-id").Value.String()

			raw, err := json.Marshal([]interface{}{map[string]interface{}{
				"op":   "remove",
				"path": "/connectorCustomizerId",
			}})
			if err != nil {
				return err
			}

			resp, err := client.Patch(cmd.Context(), util.ResourceUrl(connectorInstancesEndpoint, instanceID), bytes.NewReader(raw))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("link customizer failed. status: %s\nbody: %s", resp.Status, string(body))
			}

			var i instance
			err = json.NewDecoder(resp.Body).Decode(&i)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(instanceColumns)
			table.Append(i.columns())
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringP("instance-id", "i", "", "Connector instance ID")
	_ = cmd.MarkFlagRequired("instance-id")

	return cmd
}
