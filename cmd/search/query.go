// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/search"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newQueryCmd(folderPath string, save bool) *cobra.Command {

	var indices []string

	var sort []string
	var searchQuery string
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Manually Search using a specific Query and Indicies",
		Long:    "\nRun a search query in IdentityNow using a specific Query and Indicies\n\n",
		Example: "sail search query \"(type:provisioning AND created:[now-90d TO now])\" --indices events",
		Aliases: []string{"que"},
		Args:    cobra.ExactArgs(1),
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

			searchQuery = args[0]
			fmt.Println(searchQuery)

			searchObj, err := search.BuildSearch(searchQuery, sort, indices)
			if err != nil {
				return err
			}

			log.Info("Performing Search", "Query", searchQuery, "Indices", indices)

			formattedResponse, err := search.PerformSearch(*apiClient, searchObj)
			if err != nil {
				return err
			}

			if save {
				err = search.IterateIndices(formattedResponse, searchQuery, folderPath, []string{"json"})
				if err != nil {
					return err
				}
			} else {
				cmd.Println(util.PrettyPrint(formattedResponse))
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVar(&indices, "indices", []string{}, "indices to perform the search query on (accessprofiles, accountactivities, entitlements, events, identities, roles)")
	cmd.Flags().StringArrayVar(&sort, "sort", []string{}, "the sort value for the api call (displayName, +id...)")
	cmd.MarkFlagRequired("indices")

	return cmd
}
