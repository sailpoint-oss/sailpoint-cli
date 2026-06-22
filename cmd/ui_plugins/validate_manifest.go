// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package ui_plugins

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newValidateManifestCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "validate-manifest",
		Aliases: []string{"validate"},
		Short:   "Validate the plugin workspace manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := loadAndValidateWorkspaceManifest(manifestFileName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Manifest is valid.")
			return nil
		},
	}
}

