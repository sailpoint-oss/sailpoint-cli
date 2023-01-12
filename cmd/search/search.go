// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

const (
	searchEndpoint = "/v3/transforms"
)

func NewSearchCmd(client client.Client) *cobra.Command {
	var formats []string
	var indicies []string
	var output string
	var count bool
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search IDN with a search string",
		Long:    "Search IdentityNow with a provided search string",
		Example: "sail search \"\"",
		Aliases: []string{"se"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("search-endpoint").Value.String()
			fmt.Println(endpoint)

			if output == "" {
				output = "search_results"
			}
			fmt.Println(output)

			searchQuery := args[0]
			fmt.Println(searchQuery)

			if len(indicies) == 0 {
				indicies = []string{"accessprofiles", "accountactivities", "entitlements", "events", "identities", "roles"}
			}
			fmt.Println(indicies)

			fmt.Println(formats)

			// color.Green("Search Results saved successfully to %v", output)

			return nil
		},
	}
	cmd.Flags().BoolVarP(&count, "count", "c", false, "Return result count")
	cmd.Flags().StringArrayVarP(&formats, "formats", "f", []string{"csv"}, "Format to Save the search results")
	cmd.Flags().StringArrayVarP(&indicies, "indicies", "i", []string{}, "Indicies to search on (accessprofiles, accountactivities, entitlements, events, identities, roles)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Path to save the searchin (default current working directory).  If the directory doesn't exist, then it will be automatically created.")
	cmd.PersistentFlags().StringP("search-endpoint", "e", util.GetBasePath()+searchEndpoint, "Override search endpoint")

	return cmd

}
