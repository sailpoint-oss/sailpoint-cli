package va

import (
	"context"

	sailpoint "github.com/sailpoint-oss/golang-sdk"
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List the Clusters and Virtual Appliances configured in IdentityNow",
		Long:    "\nList the Clusters and Virtual Appliances configured in IdentityNow\n\n",
		Example: "sail va list",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			clusters, resp, err := sailpoint.PaginateWithDefaults[sailpointbetasdk.ManagedCluster](apiClient.Beta.ManagedClustersApi.GetManagedClusters(context.TODO()))
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			cmd.Println(util.PrettyPrint(clusters))

			return nil
		},
	}

	return cmd
}
