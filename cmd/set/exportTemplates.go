package set

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/spf13/cobra"
)

func newExportTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exportTemplates",
		Short:   "configure the custom export template file path",
		Long:    "configure the custom export template file path",
		Example: "sail set export full/path/to/export/templates",
		Aliases: []string{"export"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			filePath := args[0]
			if filePath == "" {
				log.Log.Error("File Path Cannot Be Blank")
			}

			config.SetCustomExportTemplatePath(filePath)

			return nil
		},
	}
	return cmd

}
