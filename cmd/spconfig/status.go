// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"

	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newExportStatusCmd(apiClient *sailpoint.APIClient) *cobra.Command {
	var exportJobs []string
	var importJobs []string
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "begin an export job in identitynow",
		Long:    "initiate an export job in identitynow",
		Example: "sail spconfig export",
		Aliases: []string{"que"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			for i := 0; i < len(exportJobs); i++ {
				job := exportJobs[i]
				ctx := context.TODO()

				status, _, err := apiClient.Beta.SPConfigApi.SpConfigExportJobStatus(ctx, job).Execute()
				if err != nil {
					return err
				}
				util.PrintJob(*status)
			}

			for i := 0; i < len(importJobs); i++ {
				job := importJobs[i]
				ctx := context.TODO()

				status, _, err := apiClient.Beta.SPConfigApi.SpConfigImportJobStatus(ctx, job).Execute()
				if err != nil {
					return err
				}
				util.PrintJob(*status)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&importJobs, "import jobs", "i", []string{}, "a list of import job ids to check the status of")
	cmd.Flags().StringArrayVarP(&exportJobs, "export jobs", "e", []string{}, "a list of export job ids to check the status of")

	return cmd
}
