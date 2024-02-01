// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	"context"
	_ "embed"
	"os"
	"strings"

	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed update.md
var updateHelp string

func newUpdateCommand() *cobra.Command {
	help := util.ParseHelp(updateHelp)
	var file bool
	var directory bool
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update a workflow in IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"up"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			var workflowFiles []string
			var workflowList []beta.Workflow

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			if directory {
				for _, workflowDirectory := range args {
					files, err := os.ReadDir(workflowDirectory)
					if err != nil {
						return err
					}

					for _, file := range files {
						if !file.IsDir() && strings.Contains(file.Name(), ".json") {
							workflowFiles = append(workflowFiles, file.Name())
						}
					}
				}
			} else if file {
				workflowFiles = args
			} else {
				cmd.Help()
				return nil
			}

			for _, workflowFile := range workflowFiles {
				var workflow beta.Workflow
				contents, err := os.ReadFile(workflowFile)
				if err != nil {
					return err
				}
				workflow.UnmarshalJSON(contents)
				workflowList = append(workflowList, workflow)
			}

			for _, workflowEntry := range workflowList {

				body, err := workflowEntry.MarshalJSON()
				if err != nil {
					return err
				}

				workFlowBody := beta.WorkflowBody{}
				workFlowBody.UnmarshalJSON(body)

				returnedWorkflow, resp, sdkErr := apiClient.Beta.WorkflowsAPI.UpdateWorkflow(context.TODO(), *workflowEntry.Id).WorkflowBody(workFlowBody).Execute()
				if sdkErr != nil {
					err := sdk.HandleSDKError(resp, sdkErr)
					if err != nil {
						return err
					}
				}

				cmd.Println(util.PrettyPrint(returnedWorkflow))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&file, "file", "f", false, "Read workflow from file(s).")
	cmd.Flags().BoolVarP(&directory, "directory", "d", false, "Read workflow from stdin.")
	cmd.MarkFlagsMutuallyExclusive("file", "directory")

	return cmd

}
