// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package ui_plugins

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a UI plugin workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("`sail ui-plugins init` is not implemented yet")
		},
	}
}
