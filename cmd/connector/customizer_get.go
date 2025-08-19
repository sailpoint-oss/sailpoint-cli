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

func newCustomizerGetCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get connector customizer",
		Example: "sail conn customizers update -c 1234",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := cmd.Flags().Lookup("id").Value.String()

			resp, err := client.Get(cmd.Context(), util.ResourceUrl(connectorCustomizersEndpoint, id), nil)
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
			table.Header(toAny(customizerColumns)...)
			table.Append(cus.columns())
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Connector customizer ID")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
