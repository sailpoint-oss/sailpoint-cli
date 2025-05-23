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

func newCustomizerCreateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <customizer-name>",
		Short:   "Create connector customizer",
		Example: "sail conn customizers create \"My Customizer\"",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := json.Marshal(customizer{Name: args[0]})
			if err != nil {
				return err
			}

			resp, err := client.Post(cmd.Context(), util.ResourceUrl(connectorCustomizersEndpoint), "application/json", bytes.NewReader(raw), nil)
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

	return cmd
}
