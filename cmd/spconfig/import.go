// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"

	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/sdk-output/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/spf13/cobra"
)

func newImportCommand() *cobra.Command {
	var payload *sailpointbetasdk.ImportOptions
	cmd := &cobra.Command{
		Use:     "import",
		Short:   "begin an import job in identitynow",
		Long:    "initiate an import job in identitynow",
		Example: "sail spconfig import",
		Aliases: []string{"que"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient := config.InitAPIClient()

			ctx := context.TODO()

			payload = sailpointbetasdk.NewImportOptions()

			job, _, err := apiClient.Beta.SPConfigApi.SpConfigImport(ctx).Data(args[0]).Options(*payload).Execute()
			if err != nil {
				return err
			}

			spconfig.PrintJob(*job)

			return nil
		},
	}

	return cmd
}
