package set

import (
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDebugCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "debug",
		Short:   "Enable/Disable debug mode.",
		Long:    "Enable or Disable debug mode for the CLI.",
		Example: "sail debug enable | disable",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			switch strings.ToLower(args[0]) {
			case "enable":
				viper.Set("debug", true)
				log.Log.Info("Debug Enabled")
			case "disable":
				viper.Set("debug", false)
				log.Log.Info("Debug Disabled")
			default:
				log.Log.Error("Invalid Selection")
			}

			return nil
		},
	}
	return cmd

}
