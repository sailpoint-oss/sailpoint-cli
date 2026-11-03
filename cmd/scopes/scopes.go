package scopes

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed scopes.md
var scopesHelp string

func NewScopesCommand() *cobra.Command {
	help := util.ParseHelp(scopesHelp)
	cmd := &cobra.Command{
		Use:     "scopes",
		Short:   "Find the OAuth scopes that your SailPoint SDK code needs",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newAnalyzeCommand(),
	)

	return cmd
}
