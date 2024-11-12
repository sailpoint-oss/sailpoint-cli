package jsonpath

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed jsonpath.md
var jsonpathHelp string

func NewJSONPathCmd() *cobra.Command {
	help := util.ParseHelp(jsonpathHelp)
	cmd := &cobra.Command{
		Use:     "jsonpath",
		Short:   "JSONPath validation for workflows and event triggers",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"jp"},
		Args:    cobra.MaximumNArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// log.Info("Hello World!")

	return cmd
}
