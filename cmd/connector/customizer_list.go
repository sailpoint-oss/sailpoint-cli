// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
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

func newCustomizerListCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all customizers",
		Example: "sail conn customizers list",
		RunE: func(cmd *cobra.Command, args []string) error {

			resp, err := client.Get(cmd.Context(), util.ResourceUrl(connectorCustomizersEndpoint))
			if err != nil {
				return err
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("list customizers failed. status: %s\nbody: %s", resp.Status, string(body))
			}

			var customizers []customizer
			err = json.NewDecoder(resp.Body).Decode(&customizers)
			if err != nil {
				return err
			}

			// raw, err := io.ReadAll(resp.Body)
			// if err != nil {
			// 	return err
			// }

			// var tags []tag
			// err = json.Unmarshal(raw, &tags)
			// if err != nil {
			// 	return err
			// }

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(customizerColumns)
			for _, c := range customizers {
				table.Append(c.columns())
			}
			table.Render()

			return nil
		},
	}

	return cmd
}
