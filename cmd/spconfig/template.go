// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
		Short:   "Begin an SPConfig Export task in IdentityNow using a template",
		Long:    "\nBegin an SPConfig Export task in IdentityNow using a template\n\n",
		Example: "sail spconfig template --wait",
		Aliases: []string{"temp"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
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

			config.Log.Info("Template Selected", "Template", template)

			matches := types.Filter(exportTemplates, func(st templates.ExportTemplate) bool { return st.Name == template })
			if len(matches) < 1 {
				return fmt.Errorf("no template matches for %s", template)
			} else if len(matches) > 1 {
				config.Log.Warn("Multiple template matches", "Template", template)
			}
			selectedTemplate = matches[0]
			varCount := len(selectedTemplate.Variables)
			if varCount > 0 {
				for i := 0; i < varCount; i++ {
					varEntry := selectedTemplate.Variables[i]
					resp := terminal.InputPrompt("Input " + varEntry.Prompt + ":")
					selectedTemplate.Raw = []byte(strings.ReplaceAll(string(selectedTemplate.Raw), "{{"+varEntry.Name+"}}", resp))
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
				config.Log.Info("Checking Export Job", "JobID", job.JobId)
				spconfig.DownloadExport(job.JobId, "spconfig-export-"+template+"-"+job.JobId+".json", folderPath)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&outputTypes, "outputTypes", "o", []string{"json"}, "the sort value for the api call (examples)")
	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "spconfig-exports", "folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for the export job to finish, and download the results")

	return cmd
}
