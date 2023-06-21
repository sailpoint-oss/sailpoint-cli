// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Initialize SDK projects",
		Long:    "\nInitialize SDK projects\n\n",
		Example: "sail sdk init",
		Aliases: []string{"temp"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newTypescriptCmd(),
		newGolangCmd(),
		newPowerShellCmd(),
		newConfigCmd(),
	)

	return cmd
}
