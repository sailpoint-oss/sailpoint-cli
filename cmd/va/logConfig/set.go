package logConfig

import (
	"context"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newSetCommand() *cobra.Command {
	var level string
	var durationInMinutes int32
	var connectors []string
	var expiration string
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Set a Virtual Appliances log configuration",
		Long:    "\nSet a Virtual Appliances log configuration\n\nA list of Connectors can be found here:\nhttps://community.sailpoint.com/t5/IdentityNow-Articles/Enabling-Connector-Logging-in-IdentityNow/ta-p/188107\n\n",
		Example: "sail va log set",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			rootLevel := beta.StandardLevel(level)

			if !rootLevel.IsValid() {
				log.Fatal("logLevel provided is invalid", "level", level)
			}

			if durationInMinutes < 5 || durationInMinutes > 1440 {
				log.Fatal("durationInMinutes provided is invalid", "durationInMinutes", durationInMinutes)
			}

			var output []beta.ClientLogConfiguration

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			logLevels := make(map[string]beta.StandardLevel)
			for j := 0; j < len(connectors); j++ {
				connector := connectors[j]
				parts := strings.Split(connector, "=")
				conLevel := beta.StandardLevel(parts[1])
				if conLevel.IsValid() {
					logLevels[parts[0]] = conLevel
				} else {
					log.Warn("Log Level Invalid", "Connector", parts[0], "LogLevel", parts[1])
				}
			}

			logConfig := beta.NewClientLogConfiguration(durationInMinutes, rootLevel)
			logConfig.LogLevels = &logLevels

			for _, clusterId := range args {

				configuration, resp, err := apiClient.Beta.ManagedClustersApi.PutClientLogConfiguration(context.TODO(), clusterId).ClientLogConfiguration(*logConfig).Execute()
				if err != nil {
					return sdk.HandleSDKError(resp, err)
				}

				output = append(output, *configuration)
			}

			cmd.Println(util.PrettyPrint(output))

			return nil
		},
	}

	cmd.Flags().StringVarP(&level, "rootLogLevel", "r", "", "Root Log Level for the log configuration")
	cmd.Flags().Int32VarP(&durationInMinutes, "durationInMinutes", "d", 30, "Duration in minutes for the log configuration.\nProvided value must be above 5 and below 1440")
	cmd.Flags().StringVarP(&expiration, "expiration", "e", "", "Expiration string value for the log configuration. Example: 2020-12-15T19:13:36.079Z")
	cmd.Flags().StringArrayVarP(&connectors, "connector", "c", []string{}, "Connectors and Log Level to configure. Example:\n-c sailpoint.connector.ADLDAPConnector=TRACE\n--connector sailpoint.connector.ADLDAPConnector=TRACE")
	cmd.MarkFlagsMutuallyExclusive("expiration", "durationInMinutes")
	return cmd
}
