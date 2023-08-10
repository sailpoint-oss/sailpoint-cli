package cluster

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/cluster/logConfig"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed cluster.md
var clusterHelp string

func NewClusterCommand() *cobra.Command {
	help := util.ParseHelp(clusterHelp)
	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   "Manage Clusters in IdentityNow",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"cl"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(),
		logConfig.NewLogCommand(),
	)

	return cmd
}
