package logConfig

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed set.md
var setHelp string

func newSetCommand() *cobra.Command {
	help := util.ParseHelp(setHelp)
	var level string
	var durationInMinutes int32
	var connectors []string
	var expiration string
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Set a Virtual Appliances log configuration",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			rootLevel := beta.StandardLevel(level)

			if !rootLevel.IsValid() {
				return errors.New("invalid logLevel: " + level)

			}

			if durationInMinutes < 5 || durationInMinutes > 1440 {
				return errors.New("invalid durationInMinutes: " + string(durationInMinutes))
			}

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			logLevels := make(map[string]beta.StandardLevel)

			for _, connector := range connectors {
				parts := strings.Split(connector, "=")
				conLevel := beta.StandardLevel(parts[1])
				if conLevel.IsValid() {
					logLevels[parts[0]] = conLevel
				} else {
					log.Warn("Log Level Invalid", "Connector", parts[0], "LogLevel", parts[1])
				}
			}

			for _, clusterId := range args {

				configuration, resp, err := apiClient.Beta.ManagedClustersApi.PutClientLogConfiguration(context.TODO(), clusterId).ClientLogConfiguration(beta.ClientLogConfiguration{DurationMinutes: durationInMinutes, RootLevel: rootLevel, LogLevels: &logLevels}).Execute()
				if err != nil {
					return sdk.HandleSDKError(resp, err)
				}

				cmd.Println(util.PrettyPrint(configuration))
			}

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
