// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package report

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	v3 "github.com/sailpoint-oss/golang-sdk/v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/templates"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/cobra"
)

func newTemplateCmd() *cobra.Command {
	var outputTypes []string
	var folderPath string
	var template string
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "generate a report using a template",
		Long:    "generate a report from IdentityNow using a template",
		Example: "sail report template",
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

			config.Log.Info("Selected Template", "template", template)

			matches := types.Filter(reportTemplates, func(st templates.ReportTemplate) bool { return st.Name == template })
			if len(matches) < 1 {
				return fmt.Errorf("no template matches for %s", template)
			} else if len(matches) > 1 {
				config.Log.Warn("multiple template matches, the first match will be used", "template", template)
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

			for i := 0; i < len(selectedTemplate.Queries); i++ {
				currentQuery := selectedTemplate.Queries[i]
				fmt.Println(currentQuery.QueryTitle + ": " + currentQuery.ResultCount)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&outputTypes, "output", "o", []string{"json"}, "the sort value for the api call (examples)")
	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "reports", "folder path to save the reports in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")

	return cmd
}
