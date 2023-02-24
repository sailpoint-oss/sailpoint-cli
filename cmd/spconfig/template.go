// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/sailpoint-oss/sailpoint-cli/internal/templates"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/cobra"
)

func newTemplateCmd() *cobra.Command {
	var outputTypes []string
	var folderPath string
	var template string
	var wait bool
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "begin an export task using a template",
		Long:    "begin an export task in IdentityNow using a template",
		Example: "sail spconfig template",
		Aliases: []string{"temp"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			err := config.InitConfig()
			if err != nil {
				return err
			}

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			if folderPath == "" {
				folderPath = "search_results"
			}

			var selectedTemplate templates.ExportTemplate
			exportTemplates, err := templates.GetExportTemplates()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				template = args[0]
			} else {
				template, err = templates.SelectTemplate(exportTemplates)
				if err != nil {
					return err
				}
			}
			if template == "" {
				return fmt.Errorf("no template specified")
			}

			color.Blue("Selected Template: %s\n", template)

			matches := types.Filter(exportTemplates, func(st templates.ExportTemplate) bool { return st.Name == template })
			if len(matches) < 1 {
				return fmt.Errorf("no template matches for %s", template)
			} else if len(matches) > 1 {
				color.Yellow("multiple template matches for %s", template)
			}
			selectedTemplate = matches[0]
			varCount := len(selectedTemplate.Variables)
			if varCount > 0 {
				for i := 0; i < varCount; i++ {
					varEntry := selectedTemplate.Variables[i]
					resp := terminal.InputPrompt(fmt.Sprintf("Input %s:", varEntry.Prompt))
					selectedTemplate.Raw = []byte(strings.ReplaceAll(string(selectedTemplate.Raw), fmt.Sprintf("{{%s}}", varEntry.Name), resp))
				}
				err := json.Unmarshal(selectedTemplate.Raw, &selectedTemplate.ExportBody)
				if err != nil {
					return err
				}
			}

			job, _, err := apiClient.Beta.SPConfigApi.ExportSpConfig(context.TODO()).ExportPayload(selectedTemplate.ExportBody).Execute()
			if err != nil {
				return err
			}

			spconfig.PrintJob(*job)

			if wait {
				color.Blue("Checking Export Job: %s", job.JobId)
				spconfig.DownloadExport(job.JobId, "spconfig-export-"+template+job.JobId+".json", folderPath)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&outputTypes, "output types", "o", []string{"json"}, "the sort value for the api call (examples)")
	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "spconfig-exports", "folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for the export job to finish, and download the results")

	return cmd
}
