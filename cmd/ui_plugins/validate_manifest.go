// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package ui_plugins

import (
	_ "embed"
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed validate_manifest.md
var validateManifestHelp string

func newValidateManifestCommand() *cobra.Command {
	help := util.ParseHelp(validateManifestHelp)
	return &cobra.Command{
		Use:     "validate-manifest",
		Aliases: []string{"validate"},
		Short:   "Validate sp-ui-plugin.json structure (offline; does not call UMS)",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := loadAndValidateWorkspaceManifest(manifestFileName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Manifest structure is valid (offline check only).")
			return nil
		},
	}
}
