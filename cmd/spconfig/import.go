// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"
	"os"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/spf13/cobra"
)

func newImportCommand() *cobra.Command {
	var filePath string
	var folderPath string
	var wait bool

	cmd := &cobra.Command{
		Use:     "import",
		Short:   "Start an Import job in IdentityNow",
		Long:    "\nStart an Import job in IdentityNow\n\n",
		Example: "sail spconfig import",
		Aliases: []string{"imp"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			ctx := context.TODO()

			job, _, err := apiClient.Beta.SPConfigApi.ImportSpConfig(ctx).Data(file).Execute()
			if err != nil {
				return err
			}

			spconfig.PrintJob(*job)

			if wait {
				log.Warn("Waiting for import task to complete")
				downloadErr := spconfig.DownloadImport(*apiClient, job.JobId, "spconfig-import-"+job.JobId, folderPath)
				if downloadErr != nil {
					return downloadErr
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "filePath", "f", "", "the path to the file containing the import payload")
	cmd.Flags().StringVarP(&folderPath, "folderPath", "p", "spconfig-imports", "folder path to save the import results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for the import job to finish, and download the results")
	cmd.MarkFlagRequired("filepath")

	return cmd
}
