package va

import (
	"context"
	_ "embed"

	sailpoint "github.com/sailpoint-oss/golang-sdk"
	"github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed list.md
var listHelp string

func newListCommand() *cobra.Command {
	help := util.ParseHelp(listHelp)
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List the Virtual Appliances configured in IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			clusters, resp, clustersErr := sailpoint.PaginateWithDefaults[beta.ManagedCluster](apiClient.Beta.ManagedClustersApi.GetManagedClusters(context.TODO()))
			if clustersErr != nil {
				return sdk.HandleSDKError(resp, clustersErr)
			}

			for _, cluster := range clusters {
				for _, id := range cluster.ClientIds {
					clientStatus, resp, clientErr := apiClient.Beta.ManagedClientsApi.GetManagedClientStatus(context.TODO(), id).Type_("VA").Execute()
					if clientErr != nil {
						return sdk.HandleSDKError(resp, clientErr)
					}
					cmd.Println(util.PrettyPrint(clientStatus))
				}
			}

			return nil
		},
	}

	return cmd
}