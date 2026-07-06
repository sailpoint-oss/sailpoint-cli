package va

import (
	"context"
	_ "embed"

	sailpoint "github.com/sailpoint-oss/golang-sdk/v3"
	"github.com/sailpoint-oss/golang-sdk/v3/managed_clients"
	"github.com/sailpoint-oss/golang-sdk/v3/managed_clusters"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

//go:embed get.md
var getHelp string

func newGetCommand() *cobra.Command {
	help := util.ParseHelp(getHelp)
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get a virtual appliance configuration from Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			var ClientIDs []string
			var VAs []managed_clients.Managedclientstatus

			clusters, resp, clustersErr := sailpoint.PaginateWithDefaults[managed_clusters.Managedcluster](apiClient.ManagedClustersAPI.GetManagedClustersV1(context.TODO()))
			if clustersErr != nil {
				return sdk.HandleSDKError(resp, clustersErr)
			}

			for _, cluster := range clusters {
				for _, id := range cluster.ClientIds {
					if len(args) > 0 {
						if slices.Contains(args, id) {
							ClientIDs = append(ClientIDs, id)

						}
					} else {
						ClientIDs = append(ClientIDs, id)
					}
				}
			}

			for _, id := range ClientIDs {
				clientStatus, resp, clientErr := apiClient.ManagedClientsAPI.GetManagedClientStatusV1(context.TODO(), id).Type_(managed_clients.MANAGEDCLIENTTYPE_VA).Execute()
				if clientErr != nil {
					return sdk.HandleSDKError(resp, clientErr)
				}
				VAs = append(VAs, *clientStatus)
			}

			cmd.Println(util.PrettyPrint(VAs))

			return nil
		},
	}

	return cmd
}
