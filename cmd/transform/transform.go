// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	transformsEndpoint      = "/v3/transforms"
	previewEndpoint         = "/cc/api/user/preview"
	identityProfileEndpoint = "/v3/identity-profiles"
	userEndpoint            = "/cc/api/identity/list"
)

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
