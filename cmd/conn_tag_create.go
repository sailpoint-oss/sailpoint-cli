// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnTagCreateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create Connector Tag",
		Example: "sp conn tags create -n rc -v 10",
		RunE: func(cmd *cobra.Command, args []string) error {
			connectorRef := cmd.Flags().Lookup("id").Value.String()
			tagName := cmd.Flags().Lookup("name").Value.String()
			versionStr := cmd.Flags().Lookup("version").Value.String()

			version, err := strconv.Atoi(versionStr)
			if err != nil {
				return err
			}

			raw, err := json.Marshal(TagCreate{TagName: tagName, ActiveVersion: uint32(version)})
			if err != nil {
				return err
			}

			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()
			resp, err := client.Post(cmd.Context(), connResourceUrl(endpoint, connectorRef, "tags"), "application/json", bytes.NewReader(raw))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("create connector tag failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err = io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var t tag
			err = json.Unmarshal(raw, &t)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(tagColumns)
			table.Append(t.columns())
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "Tag name")
	_ = cmd.MarkFlagRequired("name")

	cmd.Flags().StringP("version", "v", "", "Active version of connector upload the tag points to")
	_ = cmd.MarkFlagRequired("version")

	return cmd
}
