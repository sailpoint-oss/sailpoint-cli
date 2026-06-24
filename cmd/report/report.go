// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package report

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/charmbracelet/log"
	v3 "github.com/sailpoint-oss/golang-sdk/v2/api_v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/templates"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed report.md
var reportHelp string

func NewReportCommand() *cobra.Command {
	help := util.ParseHelp(reportHelp)
	var save bool
	var folderPath string
	var template string
	cmd := &cobra.Command{
		Use:     "report",
		Short:   "Generate a report from a template",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"rep"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			var selectedTemplate templates.ReportTemplate
			reportTemplates, err := templates.GetReportTemplates()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				template = args[0]
			} else {
				template, err = templates.SelectTemplate(reportTemplates)
				if err != nil {
					return err
				}
			}
			if template == "" {
				return fmt.Errorf("no template specified")
			}

			log.Info("Selected Template", "template", template)

			matches := types.Filter(reportTemplates, func(st templates.ReportTemplate) bool { return st.Name == template })
			if len(matches) < 1 {
				return fmt.Errorf("no template matches for %s", template)
			} else if len(matches) > 1 {
				log.Warn("multiple template matches, the first match will be used", "template", template)
			}
			selectedTemplate = matches[0]

			if len(selectedTemplate.Variables) > 0 {
				for _, varEntry := range selectedTemplate.Variables {

					resp, err := tui.Input(varEntry.Prompt, "")
					if err != nil {
						return err
					}
					selectedTemplate.Raw = []byte(strings.ReplaceAll(string(selectedTemplate.Raw), "{{"+varEntry.Name+"}}", resp))
				}
				err := json.Unmarshal(selectedTemplate.Raw, &selectedTemplate.Queries)
				if err != nil {
					return err
				}
			}

			for i, currentQuery := range selectedTemplate.Queries {

				searchQuery := v3.NewSearch()
				query := v3.NewQuery()
				query.SetQuery(currentQuery.QueryString)
				searchQuery.Query = query

				resp, err := apiClient.V3.SearchAPI.SearchCount(context.TODO()).Search(*searchQuery).Execute()
				if err != nil {
					return fmt.Errorf("failed to count results for report query %d: %w", i+1, err)
				}
				if resp == nil || len(resp.Header["X-Total-Count"]) == 0 {
					return fmt.Errorf("missing X-Total-Count header for report query %d", i+1)
				}
				selectedTemplate.Queries[i].ResultCount = resp.Header["X-Total-Count"][0]
			}

			if save {
				fileName := selectedTemplate.Name + ".json"
				err := output.SaveJSONFile(selectedTemplate.Queries, fileName, folderPath)
				if err != nil {
					return err
				}

				log.Info("Report saved", "path", path.Join(folderPath, fileName))

			} else {
				cmd.Println(util.PrettyPrint(selectedTemplate.Queries))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&save, "save", "s", false, "save the report to a file")
	cmd.Flags().StringVarP(&folderPath, "folder-path", "f", "reports", "Folder path to save reports in")
	cmd.Flags().StringVar(&folderPath, "folderPath", "reports", "Deprecated: use --folder-path")
	cmd.Flags().MarkDeprecated("folderPath", "use --folder-path")
	cmd.Flags().MarkHidden("folderPath")

	return cmd

}
