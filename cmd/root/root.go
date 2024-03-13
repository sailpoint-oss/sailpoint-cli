package root

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/cluster"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/connector"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/environment"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/report"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/rule"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/sanitize"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/search"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/set"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/spconfig"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/transform"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/va"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/workflow"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "2.1.4"

//go:embed root.md
var rootHelp string

func NewRootCommand() *cobra.Command {
	help := util.ParseHelp(rootHelp)
	var env string
	var debug bool
	root := &cobra.Command{
		Use:          "sail",
		Long:         help.Long,
		Example:      help.Example,
		Version:      version,
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	t := &terminal.Term{}

	root.AddCommand(
		cluster.NewClusterCommand(),
		connector.NewConnCmd(t),
		environment.NewEnvironmentCommand(),
		report.NewReportCommand(),
		sdk.NewSDKCommand(),
		search.NewSearchCommand(),
		set.NewSetCmd(t),
		spconfig.NewSPConfigCommand(),
		transform.NewTransformCommand(),
		rule.NewRuleCommand(),
		va.NewVACommand(t),
		workflow.NewWorkflowCommand(),
		sanitize.NewSanitizeCommand(),
	)

	root.PersistentFlags().StringVarP(&env, "env", "", "", "Environment to use for SailPoint CLI commands")
	root.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Enable debug logging")
	viper.BindPFlag("activeenvironment", root.PersistentFlags().Lookup("env"))
	viper.BindPFlag("debug", root.PersistentFlags().Lookup("debug"))

	return root
}
