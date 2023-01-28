// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newTemplateCmd(apiClient *sailpoint.APIClient) *cobra.Command {
	var output string
	var template string
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "run a search using a template",
		Long:    "run a search in IdentityNow using a search template",
		Example: "sail search template",
		Aliases: []string{"temp"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if output == "" {
				output = "search_results"
			}

			var selectedTemplate types.SearchTemplate
			searchTemplates, err := util.GetSearchTemplates()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				template = args[0]
			} else {
				var prompts []types.Choice
				for i := 0; i < len(searchTemplates); i++ {
					temp := searchTemplates[i]

					var description string
					if len(temp.Variables) > 0 {
						description = fmt.Sprintf("%s - Accepts Input", temp.Description)
					} else {
						description = temp.Description
					}
					prompts = append(prompts, types.Choice{Title: temp.Name, Description: description})
				}

				intermediate, err := tui.PromptList(prompts, "Select a Template")
				if err != nil {
					return err
				}
				template = intermediate.Title
			}
			if template == "" {
				return fmt.Errorf("no template specified")
			}

			color.Blue("Selected Template: %s\n", template)

			matches := types.Filter(searchTemplates, func(st types.SearchTemplate) bool { return st.Name == template })
			if len(matches) < 1 {
				return fmt.Errorf("no template matches for %s", template)
			} else if len(matches) > 1 {
				color.Yellow("multiple template matches for %s", template)
			}
			selectedTemplate = types.Filter(searchTemplates, func(st types.SearchTemplate) bool { return st.Name == template })[0]
			varCount := len(selectedTemplate.Variables)
			if varCount > 0 {
				for i := 0; i < varCount; i++ {
					varEntry := selectedTemplate.Variables[i]
					resp := util.InputPrompt(fmt.Sprintf("Input %s:", varEntry.Prompt))
					selectedTemplate.Raw = []byte(strings.ReplaceAll(string(selectedTemplate.Raw), fmt.Sprintf("{{%s}}", varEntry.Name), resp))
				}
				err := json.Unmarshal(selectedTemplate.Raw, &selectedTemplate.SearchQuery)
				if err != nil {
					return err
				}
			}

			color.Blue("\nPerforming Search\nQuery: \"%s\"\nIndicie: %s\n\n", selectedTemplate.SearchQuery.Query.GetQuery(), selectedTemplate.SearchQuery.Indices)

			formattedResponse, err := PerformSearch(*apiClient, selectedTemplate.SearchQuery)
			if err != nil {
				return err
			}

			fileName := fmt.Sprintf("query=%s&indicie=%s.json", selectedTemplate.SearchQuery.Query.GetQuery(), selectedTemplate.SearchQuery.Indices)

			err = SaveResults(formattedResponse, fileName, output)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")

	return cmd
}
