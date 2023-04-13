// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"
	"encoding/json"
	"os"

	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/spf13/cobra"
)

func newImportCommand() *cobra.Command {
	var filePath string
	var folderPath string
	var wait bool
	var payload *sailpointbetasdk.ImportOptions
	cmd := &cobra.Command{
		Use:     "import",
		Short:   "Start an Import job in IdentityNow",
		Long:    "\nStart an Import job in IdentityNow\n\n",
		Example: "sail spconfig import",
		Aliases: []string{"imp"},
		Args:    cobra.ExactArgs(1),
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

			err = json.NewDecoder(file).Decode(&payload)
			if err != nil {
				return err
			}

			ctx := context.TODO()

			payload = sailpointbetasdk.NewImportOptions()

			job, _, err := apiClient.Beta.SPConfigApi.ImportSpConfig(ctx).Data(args[0]).Options(*payload).Execute()
			if err != nil {
				return err
			}

			spconfig.PrintJob(*job)

			if wait {
				config.Log.Warn("Waiting for import task to complete")
				spconfig.DownloadImport(job.JobId, "spconfig-import-"+job.JobId+".json", folderPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "filePath", "f", "", "the path to the file containing the import payload")
	cmd.Flags().StringVarP(&folderPath, "folderPath", "p", "spconfig-imports", "folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for the import job to finish, and download the results")
	cmd.MarkFlagRequired("filepath")

	return cmd
}
