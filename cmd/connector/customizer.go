// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnCustomizersCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customizers",
		Short: "Manage connector customizers",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newCustomizerInitCmd(),
		newCustomizerListCmd(),
		newCustomizerCreateCmd(),
		newCustomizerGetCmd(),
		newCustomizerUpdateCmd(),
		newCustomizerDeleteCmd(),
		// upload, link, and unlink still use the raw client: the SDK's
		// version-create has no file-upload param, and connector-instances
		// (link/unlink) is not available in the SDK.
		newCustomizerCreateVersionCmd(client),
		newCustomizerLinkCmd(client),
		newCustomizerUnlinkCmd(client),
	)

	return cmd
}
