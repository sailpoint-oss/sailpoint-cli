// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	"context"
	_ "embed"

	"github.com/sailpoint-oss/golang-sdk/v3/workflows"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

//go:embed get.md
var getHelp string

func newGetCommand() *cobra.Command {
	help := util.ParseHelp(getHelp)
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get workflows in Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"g"},
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			workflowList, resp, sdkErr := apiClient.WorkflowsAPI.ListWorkflowsV1(context.TODO()).Execute()
			if sdkErr != nil {
				err := sdk.HandleSDKError(resp, sdkErr)
				if err != nil {
					return err
				}
			}

			if len(args) > 0 {
				var filteredList []workflows.Workflow
				for _, workflow := range workflowList {
					if slices.Contains(args, *workflow.Id) {
						filteredList = append(filteredList, workflow)

					}
				}
				workflowList = filteredList
			}

			cmd.Println(util.PrettyPrint(workflowList))

			return nil
		},
	}

	return cmd

}
