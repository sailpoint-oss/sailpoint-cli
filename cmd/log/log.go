// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package log

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

func NewLogCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log",
		Short:   "Parse logs",
		Aliases: []string{"log"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newParseCmd(client),
	)

	return cmd
}
