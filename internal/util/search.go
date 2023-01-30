package util

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
)

func ParseIndicie(indicie string) (sailpointsdk.Index, error) {
	switch indicie {
	case "accessprofiles":
		return sailpointsdk.INDEX_ACCESSPROFILES, nil
	case "accountactivities":
		return sailpointsdk.INDEX_ACCOUNTACTIVITIES, nil
	case "entitlements":
		return sailpointsdk.INDEX_ENTITLEMENTS, nil
	case "events":
		return sailpointsdk.INDEX_EVENTS, nil
	case "identities":
		return sailpointsdk.INDEX_IDENTITIES, nil
	case "roles":
		return sailpointsdk.INDEX_ROLES, nil
	}
	return sailpointsdk.INDEX_STAR, fmt.Errorf("indicie provided is invalid")
}

func BuildSearch(searchQuery string, sort []string, indicies []string) (sailpointsdk.Search1, error) {

	search := sailpointsdk.NewSearch1()
	search.Query = sailpointsdk.NewQuery()
	search.Query.Query = &searchQuery
	search.Sort = sort
	search.Indices = []sailpointsdk.Index{}

	for i := 0; i < len(indicies); i++ {
		tempIndicie, err := ParseIndicie(indicies[i])

		if err != nil {
			return *search, err
		}

		search.Indices = append(search.Indices, tempIndicie)
	}

	return *search, nil
}

func PerformSearch(apiClient sailpoint.APIClient, search sailpointsdk.Search1) ([]byte, error) {

	ctx := context.TODO()
	resp, r, err := sailpoint.PaginateWithDefaults[map[string]interface{}](apiClient.V3.SearchApi.SearchPost(ctx).Search1(search))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}

	color.Green("Search complete, saving results")
	formatted, err := json.MarshalIndent(resp, "", " ")
	if err != nil {
		return nil, err
	}

	return formatted, nil
}

func SaveResults(formattedResponse []byte, fileName string, output string) error {
	savePath := path.Join(output, fileName)

	// Make sure the output dir exists first
	err := os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(savePath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}

	fileWriter := bufio.NewWriter(file)

	_, err = fileWriter.Write(formattedResponse)
	if err != nil {
		return err
	}

	color.Green("Search Results saved to %s", savePath)

	return nil
}
