// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSDKCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sdk",
		Short:   "Initialize or configure SDK projects",
		Long:    "\nInitialize or configure SDK projects\n\n",
		Example: "sail sdk",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newInitCmd(),
	)

	return cmd

}
