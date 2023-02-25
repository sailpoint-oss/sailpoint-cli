// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/spf13/cobra"
)

func newDownloadCmd() *cobra.Command {
	var folderPath string
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "download results of an export job from identitynow",
		Long:    "download results of an export job from identitynow",
		Example: "sail spconfig download 37a64554-bf83-4d6a-8303-e6492251806b",
		Aliases: []string{"que"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			for i := 0; i < len(args); i++ {
				jobId := args[i]
				log.Log.Info("Checking Export Job", "JobID", jobId)
				err := spconfig.DownloadExport(jobId, "spconfig-export-"+jobId+".json", folderPath)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "spconfig-exports", "folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")

	return cmd
}
