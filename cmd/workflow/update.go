// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	"context"
	_ "embed"
	"os"
	"strings"

	"github.com/sailpoint-oss/golang-sdk/v3/workflows"
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
		Short:   "Update a workflow in Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"up"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			var workflowFiles []string
			var workflowList []workflows.Workflow

			apiClient, err := config.InitAPIClient(false)
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
				var workflow workflows.Workflow
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

				workFlowBody := workflows.Workflowbody{}
				workFlowBody.UnmarshalJSON(body)

				returnedWorkflow, resp, sdkErr := apiClient.WorkflowsAPI.PutWorkflowV1(context.TODO(), *workflowEntry.Id).Workflowbody(workFlowBody).Execute()
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
