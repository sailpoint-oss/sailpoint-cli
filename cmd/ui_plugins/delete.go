// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package ui_plugins

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a UI plugin instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("`sail ui-plugins delete` is not implemented yet")
		},
	}
}
