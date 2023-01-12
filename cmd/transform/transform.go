// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

const (
	transformsEndpoint      = "/v3/transforms"
	previewEndpoint         = "/cc/api/user/preview"
	identityProfileEndpoint = "/v3/identity-profiles"
	userEndpoint            = "/cc/api/identity/list"
)

func NewTransformCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transform",
		Short:   "Manage transforms",
		Aliases: []string{"tran"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.PersistentFlags().StringP("transforms-endpoint", "e", util.GetBasePath()+transformsEndpoint, "Override transforms endpoint")

	cmd.AddCommand(
		newListCmd(client),
		newDownloadCmd(client),
		newCreateCmd(client),
		newUpdateCmd(client),
		newDeleteCmd(client),
		newPreviewCmd(client),
	)

	return cmd
}
