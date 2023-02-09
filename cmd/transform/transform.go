// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/sdk-output/v3"
	transmodel "github.com/sailpoint-oss/sailpoint-cli/cmd/transform/model"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	transformsEndpoint      = "/v3/transforms"
	previewEndpoint         = "/cc/api/user/preview"
	identityProfileEndpoint = "/v3/identity-profiles"
	userEndpoint            = "/cc/api/identity/list"
)

func GetTransforms() ([]sailpointsdk.Transform, error) {
	apiClient := config.InitAPIClient()
	transforms, _, err := sailpoint.PaginateWithDefaults[sailpointsdk.Transform](apiClient.V3.TransformsApi.GetTransformsList(context.TODO()))
	if err != nil {
		return nil, err
	}

	return transforms, nil
}

func ListTransforms() error {

	transforms, err := GetTransforms()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(transmodel.TransformColumns)
	for _, v := range transforms {
		table.Append([]string{*v.Id, v.Name})
	}
	table.Render()

	return nil
}

func NewTransformCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transform",
		Short:   "manage transforms",
		Aliases: []string{"tran"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.PersistentFlags().StringP("transforms-endpoint", "e", transformsEndpoint, "Override transforms endpoint")

	cmd.AddCommand(
		newListCmd(),
		newDownloadCmd(),
		newCreateCmd(),
		newUpdateCmd(),
		newDeleteCmd(),
		newPreviewCmd(),
	)

	return cmd
}
