// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package ui_plugins

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUploadCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "upload",
		Short: "Upload compiled UI plugin assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("`sail ui-plugins upload` is not implemented yet")
		},
	}
}
