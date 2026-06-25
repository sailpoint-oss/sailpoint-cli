package ui_plugins

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List UI plugin instances in the current tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("`sail ui-plugins list` is not implemented yet")
		},
	}
}
