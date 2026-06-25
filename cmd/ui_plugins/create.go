package ui_plugins

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a UI plugin instance in the current tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("`sail ui-plugins create` is not implemented yet")
		},
	}
}
