// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"fmt"

	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	"github.com/spf13/cobra"
)

func NewSPConfigCmd(apiClient *sailpoint.APIClient) *cobra.Command {
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
		newExportCmd(apiClient),
		newExportStatusCmd(apiClient),
	)

	return cmd

}
