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

func newCustomizerLinkCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "link",
		Short:   "Link connector customizer to connector instance",
		Example: "sail conn customizers link -c 1234 -i 5678",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			customizerID := cmd.Flags().Lookup("id").Value.String()
			instanceID := cmd.Flags().Lookup("instance-id").Value.String()

			raw, err := json.Marshal([]interface{}{map[string]interface{}{
				"op":    "replace",
				"path":  "/connectorCustomizerId",
				"value": customizerID,
			}})
			if err != nil {
				return err
			}

			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()
			resp, err := client.Patch(cmd.Context(), util.ResourceUrl(endpoint, instanceID, "link"), bytes.NewReader(raw), nil)
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

	cmd.Flags().StringP("id", "c", "", "Connector customizer ID")
	_ = cmd.MarkFlagRequired("customizer-id")

	cmd.Flags().StringP("instance-id", "i", "", "Connector instance ID")
	_ = cmd.MarkFlagRequired("instance-id")

	cmd.Flags().StringP("conn-endpoint", "e", "", "Connector endpoint")
	_ = cmd.MarkFlagRequired("conn-endpoint")

	return cmd
}
