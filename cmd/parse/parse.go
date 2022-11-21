// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package parse

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

func NewParseCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "parse",
		Short:   "Parse logs",
		Aliases: []string{"parse"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newCCGCmd(client),
	)

	return cmd
}
