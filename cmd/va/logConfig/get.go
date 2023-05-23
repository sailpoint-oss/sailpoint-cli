package logConfig

import (
	"context"

	"github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Return a Virtual Appliances log configuration",
		Long:    "\nReturn a Virtual Appliances log configuration\n\n",
		Example: "sail va log get",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			var output []beta.ClientLogConfiguration

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			for i := 0; i < len(args); i++ {

				clusterId := args[i]

				configuration, resp, err := apiClient.Beta.ManagedClustersApi.GetClientLogConfiguration(context.TODO(), clusterId).Execute()
				if err != nil {
					return sdk.HandleSDKError(resp, err)
				}

				if configuration != nil {
					output = append(output, *configuration)
				}

			}

			cmd.Println(util.PrettyPrint(output))

			return nil
		},
	}

	return cmd
}
