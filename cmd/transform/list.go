// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/transform"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "list transforms",
		Long:    "List transforms for tenant",
		Example: "sail transform ls",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			err := transform.ListTransforms()
			if err != nil {
				return err
			}

			return nil
		},
	}
}
