// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"
	"fmt"

	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/sdk-output/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newExportCmd(apiClient *sailpoint.APIClient) *cobra.Command {
	var description string
	var includeTypes []string
	var excludeTypes []string
	var exportAll bool
	var payload *sailpointbetasdk.ExportPayload
	cmd := &cobra.Command{
		Use:     "export",
		Short:   "begin an export job in identitynow",
		Long:    "initiate an export job in identitynow",
		Example: "sail spconfig export",
		Aliases: []string{"que"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := context.TODO()

			payload = sailpointbetasdk.NewExportPayload()
			payload.Description = &description
			payload.IncludeTypes = includeTypes
			payload.ExcludeTypes = excludeTypes

			fmt.Println(payload.GetIncludeTypes())

			job, _, err := apiClient.Beta.SPConfigApi.SpConfigExport(ctx).ExportPayload(*payload).Execute()
			if err != nil {
				return err
			}

			util.PrintJob(*job)

			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "optional description for the export job")
	cmd.Flags().BoolVarP(&exportAll, "export all", "a", false, "optional flag to export all items")
	cmd.Flags().StringArrayVarP(&includeTypes, "include types", "i", []string{}, "types to include in export job")
	cmd.Flags().StringArrayVarP(&excludeTypes, "exclude types", "e", []string{}, "types to exclude in export job")
	return cmd
}
