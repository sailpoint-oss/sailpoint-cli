package set

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDebugCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "debug",
		Short:   "Enable or disable Debug Mode for the CLI",
		Long:    "\nEnable or disable Debug Mode for the CLI. Primarily used for troubleshooting.\n\n",
		Example: "sail set debug disable | sail set debug enable | sail set debug true | sail set debug false",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			switch strings.ToLower(args[0]) {
			case "enable", "true":
				viper.Set("debug", true)
				log.Info("Debug enabled")
			case "disable", "false":
				viper.Set("debug", false)
				log.Info("Debug disabled")
			default:
				log.Error("Invalid selection")
			}

			return nil
		},
	}
	return cmd

}
