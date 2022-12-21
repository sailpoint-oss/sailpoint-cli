// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newDeleteCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [TRANSFORM-ID]",
		Short:   "Delete transform",
		Long:    "Delete a transform",
		Example: "sail transform d 03d5187b-ab96-402c-b5a1-40b74285d77a",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if id == "" {
				return fmt.Errorf("transform ID cannot be empty")
			}

			endpoint := cmd.Flags().Lookup("transforms-endpoint").Value.String()
			resp, err := client.Delete(cmd.Context(), util.ResourceUrl(endpoint, id), nil)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("delete transform failed. status: %s\nbody: %s", resp.Status, body)
			}

			return nil
		},
	}

	return cmd
}
