// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package report

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/log"
	v3 "github.com/sailpoint-oss/golang-sdk/v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/templates"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func NewReportCmd() *cobra.Command {
	var save bool
	var folderPath string
	var template string
	cmd := &cobra.Command{
		Use:     "report",
		Short:   "Generate a report from a template using IdentityNow search queries",
		Long:    "Generate a report from a template using IdentityNow search queries",
		Example: "sail report \"\"",
		Aliases: []string{"rep"},
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
			varCount := len(selectedTemplate.Variables)
			if varCount > 0 {
				for i := 0; i < varCount; i++ {
					varEntry := selectedTemplate.Variables[i]
					resp := terminal.InputPrompt("Input " + varEntry.Prompt + ":")
					selectedTemplate.Raw = []byte(strings.ReplaceAll(string(selectedTemplate.Raw), "{{"+varEntry.Name+"}}", resp))
				}
				err := json.Unmarshal(selectedTemplate.Raw, &selectedTemplate.Queries)
				if err != nil {
					return err
				}
			}

			for i := 0; i < len(selectedTemplate.Queries); i++ {

				currentQuery := selectedTemplate.Queries[i]

				searchQuery := v3.NewSearch()
				query := v3.NewQuery()
				query.SetQuery(currentQuery.QueryString)
				searchQuery.Query = query

				resp, err := apiClient.V3.SearchApi.SearchCount(context.TODO()).Search(*searchQuery).Execute()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", resp)
				}
				selectedTemplate.Queries[i].ResultCount = resp.Header["X-Total-Count"][0]
			}

			// for i := 0; i < len(selectedTemplate.Queries); i++ {
			// 	currentQuery := selectedTemplate.Queries[i]
			// 	fmt.Println(currentQuery.QueryTitle + ": " + currentQuery.ResultCount)
			// }

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
	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "reports", "folder path to save the reports in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")

	return cmd

}
