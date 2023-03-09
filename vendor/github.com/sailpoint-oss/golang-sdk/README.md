# golang-sdk

### Create your project

```bash
go mod init github.com/github-repo-name/projectname
```

### Create sdk.go file and copy the below code into the file

```go
package main

import (
	"context"
	"fmt"
	"os"

	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
)

func main() {

	ctx := context.TODO()
	configuration := sailpoint.NewDefaultConfiguration()
	apiClient := sailpoint.NewAPIClient(configuration)

	resp, r, err := apiClient.V3.AccountsApi.ListAccounts(ctx).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsApi.ListAccount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAccounts`: []Account
	fmt.Fprintf(os.Stdout, "First response from `AccountsApi.ListAccount`: %v\n", resp[0].Name)

}
```

### Create a configuration file or save your configuration as environment variables

You can create a local configuration file using the [CLI tool](https://github.com/sailpoint-oss/sailpoint-cli#configuration) or you can store your configuration in environment variables
 - SAIL_BASE_URL
 - SAIL_CLIENT_ID
 - SAIL_CLIENT_SECRET

### Install sdk

```bash
go mod tidy
```

### Run the example

```bash
go run sdk.go
```


### Handling Pagination

there is a built in pagination function that can be used to automatically call and collect responses from APIs that support pagination. Use the following syntax to call it:

```go
import (
	"context"
	"fmt"
	"os"

	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	// need to import the v3 library so we are aware of the sailpointsdk.Account struct
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/sdk-output/v3"
)

func main() {

	ctx := context.TODO()
	configuration := sailpoint.NewDefaultConfiguration()
	apiClient := sailpoint.NewAPIClient(configuration)

	// use the paginate function to get 1000 results instead of hitting the normal 250 limit
	resp, r, err := sailpoint.PaginateWithDefaults[sailpointsdk.Account](apiClient.V3.AccountsApi.ListAccounts(ctx))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccountsApi.ListAccount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAccounts`: []Account
	fmt.Fprintf(os.Stdout, "First response from `AccountsApi.ListAccount`: %v\n", resp[0].Name)

}

```
### See more uses of the SDK [here](./examples/sdk.go).