package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	v3 "github.com/sailpoint-oss/golang-sdk/v2/api_v3"
)

func main() {

	ctx := context.TODO()

	configuration := sailpoint.NewDefaultConfiguration()

	apiClient := sailpoint.NewAPIClient(configuration)
	configuration.HTTPClient.RetryMax = 10

	getBeta(ctx, apiClient)

	//getAllPaginatedResults(ctx, apiClient)

}

func getResults(ctx context.Context, apiClient *sailpoint.APIClient) {
	resp, r, err := apiClient.V3.AccountsAPI.ListAccounts(ctx).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.ListAccount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAccounts`: []Account
	fmt.Fprintf(os.Stdout, "First response from `AccountsAPI.ListAccount`: %v\n", resp[0].Name)
}

func getSearchResults(ctx context.Context, apiClient *sailpoint.APIClient) {
	search := v3.NewSearchWithDefaults()
	search.Indices = append(search.Indices, "identities")
	searchString := []byte(`
	{
	"indices": [
		"identities"
	],
	"query": {
		"query": "*"
	},
    "sort": [
        "-name"
    ]
	}
	  `)
	search.UnmarshalJSON(searchString)
	resp, r, err := sailpoint.PaginateSearchApi(ctx, apiClient, *search, 0, 10, 10000)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.ListAccount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `search`
	for i := 0; i < len(resp); i++ {
		fmt.Println(resp[i]["name"])
	}
}

func getTransformResults(ctx context.Context, apiClient *sailpoint.APIClient) {
	resp, r, err := apiClient.V3.TransformsAPI.ListTransforms(ctx).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TransformsAPI.GetTransformsList``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	b, _ := json.Marshal(resp[0].Attributes)
	fmt.Fprintf(os.Stdout, "First response from `TransformsAPI.GetTransformsList`: %v\n", string(b))
}

func getAllPaginatedResults(ctx context.Context, apiClient *sailpoint.APIClient) {
	resp, r, err := sailpoint.PaginateWithDefaults[v3.Account](apiClient.V3.AccountsAPI.ListAccounts(ctx))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.ListAccount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAccounts`: []Account
	fmt.Fprintf(os.Stdout, "First response from `AccountsAPI.ListAccount`: %v\n", resp[0].Name)
}

func getBeta(ctx context.Context, apiClient *sailpoint.APIClient) {
	resp, _, err := apiClient.Beta.AccountsAPI.ListAccounts(context.TODO()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsAPI.ListAccount``: %v\n", err)
		//fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAccounts`: []Account
	fmt.Fprintf(os.Stdout, "First response from `AccountsAPI.ListAccount`: %v\n", resp)
}
