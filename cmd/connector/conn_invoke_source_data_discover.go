// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.

package connector

import (
	"encoding/json"
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnInvokeSourceDataDiscoverCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "source-data-discover [--query <value>]",
		Short:   "Invoke a std:source-data:discover command",
		Example: `sail connectors invoke source-data-discover --query '{"query": "", "limit": 10}'`,
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connRuntimeClient(cmd, client)
			if err != nil {
				return err
			}

			queryRaw := cmd.Flags().Lookup("query").Value.String()
			var queryInput map[string]any
			if err := json.Unmarshal([]byte(queryRaw), &queryInput); err != nil {
				return err
			}

			_, rawResponse, err := cc.SourceDataDiscover(ctx, queryInput)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))
			return nil
		},
	}

	cmd.Flags().StringP("query", "q", "{}", "Optional - Query to filter")

	return cmd
}
