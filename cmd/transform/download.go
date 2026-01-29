// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/mrz1836/go-sanitize"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	v3 "github.com/sailpoint-oss/golang-sdk/v2/api_v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newDownloadCommand() *cobra.Command {
	var destination string
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Download all transforms from Identity Security Cloud",
		Long:    "\nDownload all transforms from Identity Security Cloud\n\n",
		Example: "sail transform download -d transform_files | sail transform dl",
		Aliases: []string{"dl"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			transforms, resp, err := sailpoint.PaginateWithDefaults[v3.TransformRead](apiClient.V3.TransformsAPI.ListTransforms(context.TODO()))
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			filenameCounts := make(map[string]int)
			for _, v := range transforms {
				baseName := sanitize.PathName(v.Name)
				filename := baseName
				if count := filenameCounts[baseName]; count > 0 {
					filename = fmt.Sprintf("%s-%d", baseName, count)
				}
				filenameCounts[baseName]++

				err := output.SaveJSONFile(v, filename, destination)
				if err != nil {
					return err
				}
			}

			log.Info("Transforms downloaded successfully", "path", destination)

			return nil
		},
	}

	cmd.Flags().StringVarP(&destination, "destination", "d", "transform_files", "Path to the directory to save the files in (default current working directory). If the directory doesn't exist, then it will be automatically created.")

	return cmd
}
