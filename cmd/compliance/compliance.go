package compliance

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed compliance.md
var complianceHelp string

func NewComplianceCommand() *cobra.Command {
	help := util.ParseHelp(complianceHelp)

	cmd := &cobra.Command{
		Use:     "compliance",
		Short:   "Collect compliance evidence and evaluate security controls",
		Long:    help.Long,
		Example: help.Example,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newCollectCommand(),
		newEvaluateCommand(),
	)

	return cmd
}
