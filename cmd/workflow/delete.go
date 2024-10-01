// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package workflow

import (
	"context"
	_ "embed"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed delete.md
var deleteHelp string

func newDeleteCommand() *cobra.Command {
	help := util.ParseHelp(deleteHelp)
	cmd := &cobra.Command{
		Use:     "delete workflowID... ",
		Short:   "Delete a workflow in Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"del"},
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			if len(args) > 0 {

				for _, id := range args {

					resp, sdkErr := apiClient.Beta.WorkflowsAPI.DeleteWorkflow(context.TODO(), id).Execute()
					if sdkErr != nil {
						err := sdk.HandleSDKError(resp, sdkErr)
						if err != nil {
							return err
						}
					}

					if resp.StatusCode == 204 {
						log.Info("Workflow deleted", "id", id)
					} else {
						log.Warn("Workflow delete failed", "id", id)
					}

				}

			} else {
				cmd.Help()
				return nil
			}

			return nil

		},
	}

	return cmd

}
