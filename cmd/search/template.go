// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/search"
	"github.com/sailpoint-oss/sailpoint-cli/internal/templates"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newTemplateCmd(folderPath string, save bool) *cobra.Command {
	var template string
	cmd := &cobra.Command{
		Use:     "template",
		Short:   "Perform Search operations in IdentityNow using a predefined search template",
		Long:    "\nPerform Search operations in IdentityNow using a predefined search template\n\n",
		Example: "sail search template",
		Aliases: []string{"temp"},
		Args:    cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			folderPath, _ := cmd.Flags().GetString("folderPath")
			if folderPath == "" {
				cmd.MarkFlagRequired("save")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			err := config.InitConfig()
			if err != nil {
				return err
			}

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

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

			log.Info("Selected Template", "template", template)

			matches := types.Filter(searchTemplates, func(st templates.SearchTemplate) bool { return st.Name == template })
			if len(matches) < 1 {
				return fmt.Errorf("no template matches for %s", template)
			} else if len(matches) > 1 {
				log.Warn("multiple template matches, the first match will be used", "template", template)
			}
			selectedTemplate = matches[0]

			if len(selectedTemplate.Variables) > 0 {
				for _, varEntry := range selectedTemplate.Variables {

					resp := terminal.InputPrompt("Input " + varEntry.Prompt + ":")
					selectedTemplate.Raw = []byte(strings.ReplaceAll(string(selectedTemplate.Raw), "{{"+varEntry.Name+"}}", resp))
				}
				err := json.Unmarshal(selectedTemplate.Raw, &selectedTemplate.SearchQuery)
				if err != nil {
					return err
				}
			}

			log.Info("Performing Search", "Query", selectedTemplate.SearchQuery.Query.GetQuery(), "Indicies", selectedTemplate.SearchQuery.Indices)

			formattedResponse, err := search.PerformSearch(*apiClient, selectedTemplate.SearchQuery)
			if err != nil {
				return err
			}

			if save {
				err = search.IterateIndices(formattedResponse, selectedTemplate.SearchQuery.Query.GetQuery(), folderPath, []string{"json"})
				if err != nil {
					return err
				}
			} else {
				cmd.Println(util.PrettyPrint(formattedResponse))
			}

			return nil
		},
	}
	return cmd
}
