package va

import (
	"context"
	_ "embed"

	sailpoint "github.com/sailpoint-oss/golang-sdk"
	"github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
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
		Short:   "List the virtual appliances configured in IdentityNow",
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

			var clients [][]string

			for _, cluster := range clusters {
				for _, id := range cluster.ClientIds {
					clientStatus, resp, clientErr := apiClient.Beta.ManagedClientsApi.GetManagedClientStatus(context.TODO(), id).Type_("VA").Execute()
					if clientErr != nil {
						return sdk.HandleSDKError(resp, clientErr)
					}

					if clientStatus.Status != "NOT_CONFIGURED" {
						clients = append(clients, []string{*cluster.Name, clientStatus.Body["internal_ip"].(string), clientStatus.Body["id"].(string)})
					}
				}
			}

			output.WriteTable(cmd.OutOrStdout(), []string{"Cluster", "IP Address", "ID"}, clients)

			return nil
		},
	}

	return cmd
}
