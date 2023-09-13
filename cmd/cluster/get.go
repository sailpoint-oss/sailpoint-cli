package cluster

import (
	"context"
	_ "embed"

	"github.com/sailpoint-oss/golang-sdk/beta"
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
		Short:   "Get a Cluster from IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"get"},
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				var output []*beta.ManagedCluster
				for _, id := range args {
					clusters, resp, clustersErr := apiClient.Beta.ManagedClustersApi.GetManagedCluster(context.TODO(), id).Execute()
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
