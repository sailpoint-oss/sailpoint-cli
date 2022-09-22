// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"fmt"

	"github.com/sailpoint-oss/sp-cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	transformsEndpoint      = "/v3/transforms"
	previewEndpoint         = "/cc/api/user/preview"
	identityProfileEndpoint = "/v3/identity-profiles"
	userEndpoint            = "/cc/api/identity/list"
)

func NewTransformCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transforms",
		Short:   "Manage transforms",
		Aliases: []string{"trans"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.PersistentFlags().StringP("transforms-endpoint", "e", viper.GetString("baseurl")+transformsEndpoint, "Override transforms endpoint")
	cmd.PersistentFlags().StringP("preview-endpoint", "", viper.GetString("baseurl")+previewEndpoint, "Override preview endpoint")
	cmd.PersistentFlags().StringP("identity-profile-endpoint", "", viper.GetString("baseurl")+identityProfileEndpoint, "Override identity profile endpoint")
	cmd.PersistentFlags().StringP("user-endpoint", "", viper.GetString("baseurl")+userEndpoint, "Override user endpoint")

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
