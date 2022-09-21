// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"fmt"

	"github.com/sailpoint-oss/sp-cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	transformsEndpoint = "/v3/transforms"
)

func NewTransformCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transforms",
		Short:   "Manage Transforms",
		Aliases: []string{"trans"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.PersistentFlags().StringP("transforms-endpoint", "e", viper.GetString("baseurl")+transformsEndpoint, "Override transforms endpoint")

	cmd.AddCommand(
		newTransformListCmd(client),
		newTransformDownloadCmd(client),
	)

	return cmd
}
