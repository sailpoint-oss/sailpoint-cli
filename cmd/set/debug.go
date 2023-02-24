package set

import (
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newDebugCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "debug",
		Short:   "enable or disable debug mode",
		Long:    "Enable/Disable debug mode.",
		Example: "sail debug enable | disable",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			err := config.InitConfig()
			if err != nil {
				return err
			}

			switch strings.ToLower(args[0]) {
			case "enable":
				viper.Set("debug", true)
			case "disable":
				viper.Set("debug", false)
			}

			err = config.SaveConfig()
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd

}
