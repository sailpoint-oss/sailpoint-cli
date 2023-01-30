// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"fmt"

	"github.com/fatih/color"
	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newQueryCmd(apiClient *sailpoint.APIClient) *cobra.Command {
	var output string
	var indicies []string
	var sort []string
	var searchQuery string
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "search using a query",
		Long:    "Run a search query in identitynow using a query",
		Example: "sail search query",
		Aliases: []string{"que"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if output == "" {
				output = "search_results"
			}

			searchQuery = args[0]
			fmt.Println(searchQuery)

			search, err := util.BuildSearch(searchQuery, sort, indicies)
			if err != nil {
				return err
			}

			color.Blue("\nPerforming Search\nQuery: \"%s\"\nIndicie: %s\n", searchQuery, indicies)

			formattedResponse, err := util.PerformSearch(*apiClient, search)
			if err != nil {
				return err
			}

			fileName := fmt.Sprintf("query=%s&indicie=%s.json", searchQuery, indicies)

			err = util.SaveResults(formattedResponse, fileName, output)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&indicies, "indicies", "i", []string{}, "indicies to perform the search query on")
	cmd.Flags().StringArrayVarP(&sort, "sort", "s", []string{}, "the sort value for the api call (examples)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")

	cmd.MarkFlagRequired("indicies")

	return cmd
}
