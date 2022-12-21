package root

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Short:   "Change active authentication mode",
		Long:    "Change Auth Mode configured (PAT, OAuth).",
		Example: "sail auth PAT | OAuth",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			switch strings.ToLower(args[0]) {
			case "pat":
				viper.Set("authtype", "PAT")
			case "oauth":
				viper.Set("authtype", "OAuth")
			}

			err := viper.WriteConfig()
			if err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					err = viper.SafeWriteConfig()
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}

			return nil
		},
	}
	return cmd

}
