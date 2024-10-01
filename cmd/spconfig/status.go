// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	var exportJobs []string
	var importJobs []string
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Get the status of SPConfig jobs in Identity Security Cloud",
		Long:    "\nGet the status of SPConfig jobs in Identity Security Cloud\n\n",
		Example: "sail spconfig status --export 2b3b68f4-cfe7-43a6-8fb0-a518c6218111",
		Aliases: []string{"stat"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			for _, jobId := range exportJobs {

				status, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExportStatus(context.TODO(), jobId).Execute()
				if err != nil {
					return err
				}
				spconfig.PrintJob(*status)
			}

			for _, jobId := range importJobs {

				status, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigImportStatus(context.TODO(), jobId).Execute()
				if err != nil {
					return err
				}
				spconfig.PrintJob(*status)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&importJobs, "import", "", []string{}, "a list of import job ids to return the status of")
	cmd.Flags().StringArrayVarP(&exportJobs, "export", "", []string{}, "a list of export job ids to return the status of")

	return cmd
}
