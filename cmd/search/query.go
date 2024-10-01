// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/search"
	"github.com/spf13/cobra"
)

func newQueryCmd() *cobra.Command {
	var indices []string
	var sort []string
	var searchQuery string
	var folderPath string
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Manually search using a specific query and indices",
		Long:    "\nRun a search query in Identity Security Cloud, using a specific query and indicies\n\n",
		Example: "sail search query \"(type:provisioning AND created:[now-90d TO now])\" --indices events",
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
			// fmt.Println(searchQuery)

			searchObj, err := search.BuildSearch(searchQuery, sort, indices)
			if err != nil {
				return err
			}

			log.Info("Performing Search", "Query", searchQuery, "Indices", indices)

			formattedResponse, err := search.PerformSearch(*apiClient, searchObj)
			if err != nil {
				return err
			}

			err = search.IterateIndices(formattedResponse, searchQuery, folderPath, []string{"json"})
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&folderPath, "folderPath", "f", "search_results", "Folder path to save the search results to. If the directory doesn't exist, then it will be created. (defaults to the current working directory)")
	cmd.Flags().StringArrayVar(&indices, "indices", []string{}, "Indices to perform the search query on (accessprofiles, accountactivities, entitlements, events, identities, roles)")
	cmd.Flags().StringArrayVar(&sort, "sort", []string{}, "The sort value for the api call (displayName, +id...)")
	cmd.MarkFlagRequired("indices")

	return cmd
}
