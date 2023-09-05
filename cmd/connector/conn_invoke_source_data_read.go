// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.

package connector

import (
	"encoding/json"
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnInvokeSourceDataReadCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "source-data-read [sourceDataKey] [--query <value>]",
		Short:   "Invoke a std:source-data:read command",
		Example: `sail connectors invoke source-data-read john.doe --query '{"query": "jane doe", "excludeItems": ["jane","doe"], "limit": 10}'`,
		Args:    cobra.RangeArgs(1, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			queryRaw := cmd.Flags().Lookup("query").Value.String()
			var queryInput map[string]any
			if err := json.Unmarshal([]byte(queryRaw), &queryInput); err != nil {
				return err
			}

			_, rawResponse, err := cc.SourceDataRead(ctx, args[0], queryInput)
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
