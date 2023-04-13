package set

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newSearchTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "searchTemplates",
		Short:   "Set the custom IdentityNow Search templates file path",
		Long:    "\nSet the custom IdentityNow Search templates file path\n\n",
		Example: "sail set search /path/to/search/templates.json",
		Aliases: []string{"search"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			filePath := args[0]
			if filePath == "" {
				config.Log.Error("File Path Cannot Be Blank")
			}

			config.SetCustomSearchTemplatePath(filePath)

			return nil
		},
	}
	return cmd

}
