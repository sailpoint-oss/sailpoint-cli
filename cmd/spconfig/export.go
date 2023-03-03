// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"

	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/sdk-output/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var folderPath string
	var description string
	var includeTypes []string
	var excludeTypes []string
	var wait bool
	var payload *sailpointbetasdk.ExportPayload
	cmd := &cobra.Command{
		Use:     "export",
		Short:   "begin an export job in identitynow",
		Long:    "initiate an export job in identitynow",
		Example: "sail spconfig export",
		Aliases: []string{"exp"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			payload = sailpointbetasdk.NewExportPayload()
			payload.Description = &description
			payload.IncludeTypes = includeTypes
			payload.ExcludeTypes = excludeTypes

			job, _, err := apiClient.Beta.SPConfigApi.ExportSpConfig(context.TODO()).ExportPayload(*payload).Execute()
			if err != nil {
				return err
			}

			spconfig.PrintJob(*job)

			if wait {
				log.Log.Warn("Waiting for export task to complete")
				spconfig.DownloadExport(job.JobId, "spconfig-export-"+job.JobId+".json", folderPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "spconfig-exports", "folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "optional description for the export job")
	cmd.Flags().StringArrayVarP(&includeTypes, "includTypes", "i", []string{}, "types to include in export job")
	cmd.Flags().StringArrayVarP(&excludeTypes, "excludeTypes", "e", []string{}, "types to exclude in export job")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for the export job to finish, and download the results")

	return cmd
}
