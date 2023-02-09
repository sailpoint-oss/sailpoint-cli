// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSPConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "spconfig",
		Short:   "perform spconfig operations in identitynow",
		Long:    "import and export items in identitynow",
		Example: "sail spconfig",
		Aliases: []string{"spconf"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newExportCmd(),
		newExportStatusCmd(),
		newTemplateCmd(),
		newDownloadCmd(),
		newImportCommand(),
	)

	return cmd

}
