// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package report

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "report",
		Short:   "Generate a report from a template using IdentityNow search queries",
		Long:    "Generate a report from a template using IdentityNow search queries",
		Example: "sail report \"\"",
		Aliases: []string{"rep"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newTemplateCmd(),
	)

	return cmd

}
