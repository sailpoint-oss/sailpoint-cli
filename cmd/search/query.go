// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/search"
	"github.com/spf13/cobra"
)

func newQueryCmd() *cobra.Command {
	var folderPath string
	var indices []string
	var outputTypes []string
	var sort []string
	var searchQuery string
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "search using a query",
		Long:    "Run a search query in identitynow using a query",
		Example: "sail search query \"attributes.location:london\"",
		Aliases: []string{"que"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			err := config.InitConfig()
			if err != nil {
				return err
			}

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			searchQuery = args[0]
			fmt.Println(searchQuery)

			searchObj, err := search.BuildSearch(searchQuery, sort, indices)
			if err != nil {
				return err
			}

			log.Log.Info("Performing Search", "Query", searchQuery, "Indices", indices)

			formattedResponse, err := search.PerformSearch(*apiClient, searchObj)
			if err != nil {
				return err
			}

			err = search.IterateIndices(formattedResponse, searchQuery, folderPath, outputTypes)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&indices, "indices", "i", []string{}, "indices to perform the search query on (accessprofiles, accountactivities, entitlements, events, identities, roles)")
	cmd.Flags().StringArrayVarP(&sort, "sort", "s", []string{}, "the sort value for the api call (displayName, +id...)")
	cmd.Flags().StringArrayVarP(&outputTypes, "outputTypes", "o", []string{"json"}, "the output types for the results (csv, json)")
	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "search_results", "folder path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")

	cmd.MarkFlagRequired("indices")

	return cmd
}
