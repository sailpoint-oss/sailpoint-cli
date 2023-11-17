// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	"context"
	_ "embed"

	"github.com/sailpoint-oss/golang-sdk/beta"
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
		Short:   "Get workflows in IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"g"},
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

			if len(args) > 0 {
				var filteredList []beta.Workflow
				for _, workflow := range workflows {
					if slices.Contains(args, *workflow.Id) {
						filteredList = append(filteredList, workflow)

					}
				}
				workflows = filteredList
			}

			cmd.Println(util.PrettyPrint(workflows))

			return nil
		},
	}

	return cmd

}
