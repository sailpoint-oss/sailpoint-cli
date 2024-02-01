package va

import (
	"context"
	_ "embed"

	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
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
		Short:   "Get a virtual appliance configuration from IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			var ClientIDs []string
			var VAs []beta.ManagedClientStatus

			clusters, resp, clustersErr := sailpoint.PaginateWithDefaults[beta.ManagedCluster](apiClient.Beta.ManagedClustersAPI.GetManagedClusters(context.TODO()))
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
				clientStatus, resp, clientErr := apiClient.Beta.ManagedClientsAPI.GetManagedClientStatus(context.TODO(), id).Type_("VA").Execute()
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
