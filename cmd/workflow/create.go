// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"
	"strings"

	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed create.md
var createHelp string

func newCreateCommand() *cobra.Command {
	help := util.ParseHelp(createHelp)
	var file bool
	var directory bool
	cmd := &cobra.Command{
		Use:     "create [-f file1 file2 ... | -d workflowDirectory ]",
		Short:   "Create workflows in IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"cr"},
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			var workflows []beta.Workflow
			var returnedWorkflows []beta.Workflow
			var workflowFiles []string

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

			for _, filePath := range workflowFiles {

				file, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
				if err != nil {
					return err
				}

				decoder := json.NewDecoder(file)
				decoder.DisallowUnknownFields()

				var workflow beta.Workflow
				err = decoder.Decode(&workflow)
				if err != nil {
					return err
				}

				workflows = append(workflows, workflow)

			}

			for _, workflow := range workflows {
				body, err := workflow.MarshalJSON()
				if err != nil {
					return err
				}

				createReq := beta.CreateWorkflowRequest{}

				err = createReq.UnmarshalJSON(body)
				if err != nil {
					return err
				}

				workflowResp, resp, sdkErr := apiClient.Beta.WorkflowsAPI.CreateWorkflow(context.TODO()).CreateWorkflowRequest(createReq).Execute()
				if sdkErr != nil {
					err := sdk.HandleSDKError(resp, sdkErr)
					if err != nil {
						return err
					}
				}

				returnedWorkflows = append(returnedWorkflows, *workflowResp)
			}

			cmd.Println(util.PrettyPrint(returnedWorkflows))

			return nil

		},
	}

	cmd.Flags().BoolVarP(&file, "file", "f", false, "Specifies that workflow file paths are provided as arguments to be created")
	cmd.Flags().BoolVarP(&directory, "directory", "d", false, "Specifies that a directory of workflows is provided to be created")
	cmd.MarkFlagsMutuallyExclusive("file", "directory")

	return cmd

}
