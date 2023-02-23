// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/search"
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
		Short:   "run a search using a template",
		Long:    "run a search in IdentityNow using a search template",
		Example: "sail search template",
		Aliases: []string{"temp"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			err := config.InitConfig()
			if err != nil {
				return err
			}

			apiClient := config.InitAPIClient()

			var selectedTemplate templates.SearchTemplate
			searchTemplates, err := templates.GetSearchTemplates()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				template = args[0]
			} else {
				template, err = templates.SelectTemplate(searchTemplates)
				if err != nil {
					return err
				}
			}
			if template == "" {
				return fmt.Errorf("no template specified")
			}

			color.Blue("Selected Template: %s\n", template)

			matches := types.Filter(searchTemplates, func(st templates.SearchTemplate) bool { return st.Name == template })
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
				err := json.Unmarshal(selectedTemplate.Raw, &selectedTemplate.SearchQuery)
				if err != nil {
					return err
				}
			}

			color.Blue("\nPerforming Search\nQuery: \"%s\"\nIndicie: %s\n\n", selectedTemplate.SearchQuery.Query.GetQuery(), selectedTemplate.SearchQuery.Indices)

			formattedResponse, err := search.PerformSearch(*apiClient, selectedTemplate.SearchQuery)
			if err != nil {
				return err
			}

			err = search.IterateIndicies(formattedResponse, selectedTemplate.SearchQuery.Query.GetQuery(), folderPath, outputTypes)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&outputTypes, "output types", "o", []string{"json"}, "the sort value for the api call (examples)")
	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "search_results", "folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")

	return cmd
}
