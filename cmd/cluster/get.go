package cluster

import (
	"context"
	_ "embed"

	"github.com/sailpoint-oss/golang-sdk/v3/managed_clusters"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed get.md
var getHelp string

func newGetCommand() *cobra.Command {
	help := util.ParseHelp(getHelp)
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Get a cluster from Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"get"},
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			if len(args) > 0 {
				var output []*managed_clusters.Managedcluster
				for _, id := range args {
					clusters, resp, clustersErr := apiClient.ManagedClustersAPI.GetManagedClusterV1(context.TODO(), id).Execute()
					if clustersErr != nil {
						return sdk.HandleSDKError(resp, clustersErr)
					}

					output = append(output, clusters)
				}
				cmd.Println(util.PrettyPrint(output))
			} else {
				cmd.Help()
			}

			return nil
		},
	}

	return cmd
}
