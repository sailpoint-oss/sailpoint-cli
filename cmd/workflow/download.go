// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	"context"
	_ "embed"
	"os"
	"path"

	clean "github.com/mrz1836/go-sanitize"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed download.md
var downloadHelp string

func newDownloadCommand() *cobra.Command {
	help := util.ParseHelp(downloadHelp)
	var folderPath string
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Download Workflows from IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"down"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			workflows, resp, sdkErr := apiClient.Beta.WorkflowsApi.ListWorkflows(context.TODO()).Execute()
			if sdkErr != nil {
				err := sdk.HandleSDKError(resp, sdkErr)
				if err != nil {
					return err
				}
			}

			for _, v := range workflows {
				fileName := clean.PathName(*v.Name) + ".json"

				fullPath := path.Join(folderPath, fileName)

				err := os.MkdirAll(folderPath, os.ModePerm)
				if err != nil {
					return err
				}

				file, err := os.Create(fullPath)
				if err != nil {
					return err
				}

				defer file.Close()

				_, err = file.WriteString(util.PrettyPrint(v))
				if err != nil {
					return err
				}

				err = file.Sync()
				if err != nil {
					return err
				}

			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&folderPath, "folder", "f", "workflows", "Folder to save the Workflows to")

	return cmd

}
