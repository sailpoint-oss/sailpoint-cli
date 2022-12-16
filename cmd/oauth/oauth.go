// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package oauth

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

func NewOauthCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "oauth",
		Short:   "Login to OAuth ",
		Aliases: []string{"oauth"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newLoginCmd(client),
	)

	return cmd
}
