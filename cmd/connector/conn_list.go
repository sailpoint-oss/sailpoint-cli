// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newConnListCmd(client client.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List Connectors",
		Long:    "List Connectors For Tenant",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()

			resp, err := client.Get(cmd.Context(), endpoint)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("non-200 response: %s\nbody: %s", resp.Status, body)
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var conns []connectorList
			err = json.Unmarshal(raw, &conns)
			if err != nil {
				return err
			}

			// Sort connectors by Alias
			sort.Slice(conns, func(i, j int) bool {
				return conns[i].Alias < conns[j].Alias
			})

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(connectorListColumns)

			// Process each connector and populate the table
			for _, conn := range conns {
				connectorRef := conn.ID
				if connectorRef == "" {
					continue
				}

				// Build the tags endpoint using the connectorRef
				tagsEndpoint := util.ResourceUrl(endpoint, connectorRef, "tags")
				tagsResp, err := client.Get(cmd.Context(), tagsEndpoint)
				if err != nil {
					return err
				}
				defer tagsResp.Body.Close()

				if tagsResp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(tagsResp.Body)
					return fmt.Errorf("non-200 response: %s\nbody: %s", tagsResp.Status, body)
				}

				// Process the response for the tags request
				tagsRaw, err := io.ReadAll(tagsResp.Body)
				if err != nil {
					return err
				}

				var tags []tag
				err = json.Unmarshal(tagsRaw, &tags)
				if err != nil {
					return err
				}

				// Prepare data for the table
				var tagNames []string
				var versions []string
				for _, t := range tags {
					tagNames = append(tagNames, t.TagName)
					versions = append(versions, fmt.Sprintf("%d", t.ActiveVersion))
				}
				tagsString := fmt.Sprintf("%s", tagNames)
				versionsString := fmt.Sprintf("%s", versions)

				// Add the row to the table
				row := []string{
					conn.ID,
					conn.Alias,
					tagsString,
					versionsString,
				}
				table.Append(row)
			}

			// Render the table
			table.Render()
			return nil
		},
	}
}
