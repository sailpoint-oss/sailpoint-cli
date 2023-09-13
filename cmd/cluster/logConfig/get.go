package logConfig

import (
	"context"
	_ "embed"

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
		Short:   "Get a Virtual Appliances clusters log configuration",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			for _, clusterId := range args {

				configuration, resp, err := apiClient.Beta.ManagedClustersApi.GetClientLogConfiguration(context.TODO(), clusterId).Execute()
				if err != nil {
					return sdk.HandleSDKError(resp, err)
				}

				cmd.Println(util.PrettyPrint(configuration))
			}

			return nil
		},
	}

	return cmd
}
