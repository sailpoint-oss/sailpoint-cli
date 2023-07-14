// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package root

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/connector"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/environment"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/report"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/search"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/set"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/spconfig"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/transform"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/va"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/spf13/cobra"
)

var version = "1.2.0"

func NewRootCmd() *cobra.Command {
	var env string
	root := &cobra.Command{
		Use:          "sail",
		Short:        "The SailPoint CLI allows you to administer your IdentityNow tenant from the command line.\nNavigate to https://developer.sailpoint.com/idn/tools/cli to learn more.",
		Version:      version,
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			var tempEnv string
			if env != "" {
				tempEnv = config.GetActiveEnvironment()
				config.SetActiveEnvironment(env)
			}

			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())

			if tempEnv != "" {
				config.SetActiveEnvironment(tempEnv)
			}
		},
	}

	t := &terminal.Term{}

	root.AddCommand(
		set.NewSetCmd(t),
		environment.NewEnvironmentCmd(),
		connector.NewConnCmd(t),
		transform.NewTransformCmd(),
		va.NewVACmd(t),
		search.NewSearchCmd(),
		spconfig.NewSPConfigCmd(),
		report.NewReportCmd(),
		sdk.NewSDKCmd(),
	)

	root.PersistentFlags().StringVar(&env, "env", "", "Environment to use for SailPoint CLI commands")

	return root
}
