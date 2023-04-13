// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package root

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/configure"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/connector"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/environment"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/report"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/search"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/set"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/spconfig"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/transform"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/va"
	"github.com/spf13/cobra"
)

var version = "1.0.0"

func NewRootCmd() *cobra.Command {
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
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	root.AddCommand(
		set.NewSetCommand(),
		environment.NewEnvironmentCommand(),
		configure.NewConfigureCmd(),
		connector.NewConnCmd(),
		transform.NewTransformCmd(),
		va.NewVACmd(),
		search.NewSearchCmd(),
		spconfig.NewSPConfigCmd(),
		report.NewReportCommand(),
	)
	return root
}
