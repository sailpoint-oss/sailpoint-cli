package root

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Short:   "change currently active authentication mode",
		Long:    "Change Auth Mode configured (pat, oauth).",
		Example: "sail auth pat | oauth",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var selection string
			var err error
			if len(args) > 0 {
				selection = args[0]
			} else {
				selection, err = PromptAuth()
				if err != nil {
					return err
				}
			}

			switch strings.ToLower(selection) {
			case "pat":
				viper.Set("authtype", "pat")
			case "oauth":
				viper.Set("authtype", "oauth")
			default:
				return errors.New("invalid selection")
			}

			err = viper.WriteConfig()
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
