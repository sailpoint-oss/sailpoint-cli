// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sp-cli/client"
	"github.com/spf13/cobra"
)

func newTransformListCmd(client client.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List Transforms",
		Long:    "List Transforms For Tenant",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("transforms-endpoint").Value.String()

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

			var transforms []transform
			err = json.Unmarshal(raw, &transforms)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(transformColumns)
			for _, v := range transforms {
				table.Append(v.transformToColumns())
			}
			table.Render()

			return nil
		},
	}
}
