package set

import (
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
)

func PromptAuth() (string, error) {
	items := []tui.Choice{
		{Title: "PAT", Description: "Person Access Token - Single User"},
		// {Title: "OAuth", Description: "OAuth2.0 Authentication - Sign in via the Web Portal"},
	}

	choice, err := tui.PromptList(items, "Choose an authentication method to configure")
	if err != nil {
		return "", err
	}

	return strings.ToLower(choice.Title), nil
}

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Short:   "change currently active authentication mode",
		Long:    "Change Auth Mode Configured (pat, pipeline).",
		Example: "sail auth pat | sail auth pat | sail auth pipeline",
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

				config.SetAuthType("pat")
				log.Log.Info("Authentication method set to PAT")

			case "oauth":
				log.Log.Error("OAuth is not currently supported")

				// 	config.SetAuthType("oauth")
				// 	log.Log.Info("Authentication method set to OAuth")

			default:
				log.Log.Error("Invalid Selection")
			}

			return nil
		},
	}
	return cmd

}
