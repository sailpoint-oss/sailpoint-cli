// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package search

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/fatih/color"
	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/sdk-output/v3"
	"github.com/spf13/cobra"
)

// const (
// 	searchEndpoint = "/v3/search"
// )

func NewSearchCmd(apiClient *sailpoint.APIClient) *cobra.Command {
	var Formats []string
	var Indicie string
	var output string
	var sort string
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "perform search in identitynow with a search string",
		Long:    "Search IdentityNow with a provided search string",
		Example: "sail search \"\"",
		Aliases: []string{"se"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if output == "" {
				output = "search_results"
			}

			searchQuery := args[0]
			color.Blue("Running search \nQuery: \"%s\"\nIndicie: %s", searchQuery, Indicie)

			search := sailpointsdk.NewSearch1()
			search.Query = sailpointsdk.NewQuery()
			search.Query.Query = &searchQuery
			search.Indices = []sailpointsdk.Index{}

			switch Indicie {
			case "accessprofiles":
				search.Indices = append(search.Indices, sailpointsdk.INDEX_ACCESSPROFILES)
			case "accountactivities":
				search.Indices = append(search.Indices, sailpointsdk.INDEX_ACCOUNTACTIVITIES)
			case "entitlements":
				search.Indices = append(search.Indices, sailpointsdk.INDEX_ENTITLEMENTS)
			case "events":
				search.Indices = append(search.Indices, sailpointsdk.INDEX_EVENTS)
			case "identities":
				search.Indices = append(search.Indices, sailpointsdk.INDEX_IDENTITIES)
			case "roles":
				search.Indices = append(search.Indices, sailpointsdk.INDEX_ROLES)
			default:
				return fmt.Errorf("provided search indicie \"%s\" is invalid", Indicie)
			}

			ctx := context.TODO()
			resp, r, err := sailpoint.PaginateWithDefaults[map[string]interface{}](apiClient.V3.SearchApi.SearchPost(ctx).Search1(*search))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			}

			color.Green("Search complete, saving results")

			formatted, err := json.MarshalIndent(resp, "", " ")
			if err != nil {
				return err
			}

			savePath := path.Join(output, fmt.Sprintf("query=%sindicie=%s.json", searchQuery, Indicie))
			fmt.Println(savePath)

			// Make sure the output dir exists first
			err = os.MkdirAll(output, os.ModePerm)
			if err != nil {
				return err
			}

			file, err := os.OpenFile(savePath, os.O_CREATE|os.O_RDWR, 0777)
			if err != nil {
				return err
			}

			fileWriter := bufio.NewWriter(file)

			_, err = fileWriter.Write(formatted)
			if err != nil {
				return err
			}

			color.Green("Search Results saved to %s", savePath)

			return nil
		},
	}
	cmd.Flags().StringArrayVarP(&Formats, "formats", "f", []string{"json"}, "formats to save the search results in")
	cmd.Flags().StringVarP(&Indicie, "indicie", "i", "", "indicie to perform the search query on")
	cmd.Flags().StringVarP(&output, "output", "o", "", "path to save the search results in. If the directory doesn't exist, then it will be automatically created. (default is the current working directory)")
	cmd.Flags().StringVarP(&sort, "sort", "s", "", "the sort value for the api call (examples)")
	cmd.MarkFlagRequired("indicie")
	// cmd.PersistentFlags().StringP("search-endpoint", "e", searchEndpoint+"?count=true&limit=250", "override search endpoint")

	return cmd

}
