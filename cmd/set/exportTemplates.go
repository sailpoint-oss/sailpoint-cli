package set

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newExportTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exportTemplates",
		Short:   "configure the custom export template file path",
		Long:    "configure the custom export template file path",
		Example: "sail set export /path/to/export/templates",
		Aliases: []string{"export"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			config.SetCustomExportTemplatePath(args[0])

			return nil
		},
	}
	return cmd

}
