package ui_plugins

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLinkCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Link local development URL for a UI plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("`sail ui-plugins link` is not implemented yet")
		},
	}
}
