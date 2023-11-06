// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	_ "embed"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed download.md
var downloadHelp string

func newDownloadCommand() *cobra.Command {
	help := util.ParseHelp(downloadHelp)
	var importIDs []string
	var exportIDs []string
	var folderPath string
	cmd := &cobra.Command{
		Use:     "download {--import <importID> --export <exportID>}",
		Short:   "Download the results of import or export jobs from IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"down"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			for _, jobId := range importIDs {
				log.Info("Checking Import Job", "JobID", jobId)
				err := spconfig.DownloadImport(*apiClient, jobId, "spconfig-import-"+jobId+".json", folderPath)
				if err != nil {
					return err
				}
			}

			for _, jobId := range exportIDs {
				log.Info("Checking Export Job", "JobID", jobId)
				err := spconfig.DownloadExport(*apiClient, jobId, "spconfig-export-"+jobId+".json", folderPath)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&importIDs, "import", "", []string{}, "Specify the IDs of the import jobs to download results for")
	cmd.Flags().StringArrayVarP(&exportIDs, "export", "", []string{}, "Specify the IDs of the export jobs to download results for")
	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "spconfig-exports", "Folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")

	return cmd
}
