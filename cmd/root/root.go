package root

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/api"
	cmdauth "github.com/sailpoint-oss/sailpoint-cli/cmd/auth"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/cluster"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/configure"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/connector"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/env"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/environment"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/jsonpath"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/reassign"
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

var version = "2.2.12"

//go:embed root.md
var rootHelp string

func NewRootCommand() *cobra.Command {
	help := util.ParseHelp(rootHelp)
	var envFlag string
	var debug bool
	var verbose bool
	var jsonOutput bool
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().Changed("env") {
				viper.Set("activeenvironment", envFlag)
			}
			if cmd.Flags().Changed("debug") || cmd.Flags().Changed("verbose") {
				viper.Set("debug", debug || verbose)
			}
			if cmd.Flags().Changed("json") {
				viper.Set("json", jsonOutput)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	t := &terminal.Term{}

	// New command structure
	root.AddCommand(
		env.NewEnvCommand(),
		cmdauth.NewAuthCommand(),
		configure.NewConfigureCommand(),
	)

	// Existing commands (unchanged)
	root.AddCommand(
		api.NewAPICommand(),
		cluster.NewClusterCommand(),
		connector.NewConnCmd(t),
		jsonpath.NewJSONPathCmd(),
		report.NewReportCommand(),
		sdk.NewSDKCommand(),
		search.NewSearchCommand(),
		spconfig.NewSPConfigCommand(),
		transform.NewTransformCommand(),
		rule.NewRuleCommand(),
		va.NewVACommand(),
		workflow.NewWorkflowCommand(),
		sanitize.NewSanitizeCommand(),
		reassign.NewReassignCommand(),
	)

	// Deprecated commands (kept for backward compatibility)
	deprecatedSet := set.NewSetCmd(t)
	deprecatedSet.Deprecated = "use 'sail config <key> <value>' for settings, 'sail env create/update' for auth configuration"
	root.AddCommand(deprecatedSet)

	deprecatedEnv := environment.NewEnvironmentCommand()
	deprecatedEnv.Deprecated = "use 'sail env' instead"
	root.AddCommand(deprecatedEnv)

	root.PersistentFlags().StringVar(&envFlag, "env", "", "Environment to use for SailPoint CLI commands")
	root.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output (same as --debug)")
	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")

	return root
}
