// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package root

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/connector"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/transform"
	"github.com/spf13/cobra"
)

var version = "0.0.1"

func NewRootCmd(client client.Client) *cobra.Command {
	root := &cobra.Command{
		Use:          "sail",
		Short:        "sail",
		Version:      version,
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}
	root.AddCommand(
		newConfigureCmd(client),
		connector.NewConnCmd(client),
		transform.NewTransformCmd(client),
	)
	return root
}
