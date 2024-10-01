// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/sailpoint-oss/sailpoint-cli/internal/templates"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed template.md
var templateHelp string

func newTemplateCommand() *cobra.Command {
	help := util.ParseHelp(templateHelp)
	var folderPath string
	var template string
	var wait bool
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "Begin an SPConfig export task in Identity Security Cloud, using a template",
		Long:    help.Long,
		Example: help.Example,
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

			log.Info("Template Selected", "Template", template)

			matches := types.Filter(exportTemplates, func(st templates.ExportTemplate) bool { return st.Name == template })
			if len(matches) < 1 {
				return fmt.Errorf("no template matches for %s", template)
			} else if len(matches) > 1 {
				log.Warn("Multiple template matches", "Template", template)
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

			job, _, err := apiClient.Beta.SPConfigAPI.ExportSpConfig(context.TODO()).ExportPayload(selectedTemplate.ExportBody).Execute()
			if err != nil {
				return err
			}

			spconfig.PrintJob(*job)

			if wait {
				log.Info("Checking Export Job", "JobID", job.JobId)
				spconfig.DownloadExport(*apiClient, job.JobId, "spconfig-export-"+template+"-"+job.JobId+".json", folderPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "spconfig-exports", "Folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")
	cmd.Flags().BoolVarP(&wait, "wait", "w", false, "Wait for the export job to finish, and then download the results")

	return cmd
}
