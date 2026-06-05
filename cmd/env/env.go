package env

import (
	"sort"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage CLI environments",
		Long:  "\nManage SailPoint Identity Security Cloud environments for the CLI.\n\nEach environment represents a tenant with its own authentication configuration.\n",
		Example: `  sail env list
  sail env create production
  sail env use staging
  sail env show`,
		// "environment" alias not used here since the deprecated cmd/environment still exists
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(),
		newShowCommand(),
		newCreateCommand(),
		newUpdateCommand(),
		newDeleteCommand(),
		newUseCommand(),
	)

	return cmd
}

func completeEnvironmentNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	environments := config.GetEnvironments()
	names := make([]string, 0, len(environments))
	for name := range environments {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}
