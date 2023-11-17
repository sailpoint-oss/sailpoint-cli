// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/golang-sdk/beta"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed export.md
var exportHelp string

func newExportCommand() *cobra.Command {
	help := util.ParseHelp(exportHelp)

	var objectOptions string
	var folderPath string
	var description string
	var includeTypes []string
	var excludeTypes []string
	var wait bool

	cmd := &cobra.Command{
		Use:     "export",
		Short:   "Start an export job in IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"exp"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			var options *map[string]beta.ObjectExportImportOptions

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			if objectOptions != "" {
				err = json.Unmarshal([]byte(objectOptions), &options)
				if err != nil {
					return err
				}
			}

			job, _, err := apiClient.Beta.SPConfigApi.ExportSpConfig(context.TODO()).ExportPayload(beta.ExportPayload{Description: &description, IncludeTypes: includeTypes, ExcludeTypes: excludeTypes, ObjectOptions: options}).Execute()
			if err != nil {
				return err
			}

			spconfig.PrintJob(*job)

			if wait {
				log.Warn("Waiting for export task to complete")
				downloadErr := spconfig.DownloadExport(*apiClient, job.JobId, "spconfig-export-"+job.JobId, folderPath)
				if downloadErr != nil {
					return downloadErr
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "spconfig-exports", "Folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")
	cmd.Flags().StringVarP(&description, "description", "", "", "Optional description for the export job")
	cmd.Flags().StringArrayVarP(&includeTypes, "include", "i", []string{}, "Types to include in export job")
	cmd.Flags().StringArrayVarP(&excludeTypes, "exclude", "e", []string{}, "Types to exclude in export job")
	cmd.Flags().StringVarP(&objectOptions, "objectOptions", "o", "", "Options for the object types being exported")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "Wait for the export job to finish, and then download the results")

	return cmd
}
