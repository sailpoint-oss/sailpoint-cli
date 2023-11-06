package set

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newExportTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exportTemplates",
		Short:   "Set the custom SPConfig export templates file path",
		Long:    "\nSet the custom SPConfig export templates file path\n\n",
		Example: "sail set export full/path/to/export/templates.json",
		Aliases: []string{"export"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			filePath := args[0]
			if filePath == "" {
				log.Error("File path cannot be blank")
			}

			config.SetCustomExportTemplatePath(filePath)

			return nil
		},
	}
	return cmd

}
