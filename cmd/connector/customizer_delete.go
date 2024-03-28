// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

func newCustomizerDeleteCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete connector customizer",
		Example: "sail conn customizers delete -c 1234",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := cmd.Flags().Lookup("id").Value.String()

			resp, err := client.Delete(cmd.Context(), util.ResourceUrl(connectorCustomizersEndpoint, id))
			if err != nil {
				return err
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("delete customizer failed. status: %s\nbody: %s", resp.Status, string(body))
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "connector customizer %s deleted.\n", id)
			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Connector customizer ID")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
