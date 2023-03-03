// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/spf13/cobra"
)

func newExportStatusCmd() *cobra.Command {
	var exportJobs []string
	var importJobs []string
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "begin an export job in identitynow",
		Long:    "initiate an export job in identitynow",
		Example: "sail spconfig export",
		Aliases: []string{"stat"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			for i := 0; i < len(exportJobs); i++ {
				job := exportJobs[i]

				status, _, err := apiClient.Beta.SPConfigApi.ExportSpConfigJobStatus(context.TODO(), job).Execute() //SPConfigApi.SpConfigExportJobStatus(ctx, job).Execute()
				if err != nil {
					return err
				}
				spconfig.PrintJob(*status)
			}

			for i := 0; i < len(importJobs); i++ {
				job := importJobs[i]

				status, _, err := apiClient.Beta.SPConfigApi.ImportSpConfigJobStatus(context.TODO(), job).Execute()
				if err != nil {
					return err
				}
				spconfig.PrintJob(*status)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&importJobs, "import jobs", "i", []string{}, "a list of import job ids to check the status of")
	cmd.Flags().StringArrayVarP(&exportJobs, "export jobs", "e", []string{}, "a list of export job ids to check the status of")

	return cmd
}
