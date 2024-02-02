// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	"context"
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed list.md
var listHelp string

func newListCommand() *cobra.Command {
	help := util.ParseHelp(listHelp)
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all workflows in IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			workflows, resp, sdkErr := apiClient.Beta.WorkflowsAPI.ListWorkflows(context.TODO()).Execute()
			if sdkErr != nil {
				err := sdk.HandleSDKError(resp, sdkErr)
				if err != nil {
					return err
				}
			}

			var tableList [][]string
			for _, entry := range workflows {
				tableList = append(tableList, []string{*entry.Name, *entry.Id})
			}

			output.WriteTable(cmd.OutOrStdout(), []string{"Name", "ID"}, tableList)

			return nil
		},
	}

	return cmd

}
