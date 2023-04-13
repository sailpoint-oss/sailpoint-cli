package set

import (
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
)

func PromptAuth() (string, error) {
	items := []tui.Choice{
		{Title: "PAT", Description: "Person Access Token - Single User PAT Configuration"},
		{Title: "OAuth", Description: "OAuth2.0 Authentication - Sign in via the IdentityNow Web Portal"},
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
		Short:   "Set the currently active authentication mode (PAT, OAuth)",
		Long:    "\nSet the currently active authentication mode\n\nSupported Authentication Methods:\nPAT\nOAuth",
		Example: "sail set auth pat | sail set auth oauth",
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
				config.Log.Info("Authentication method set to PAT")

			case "oauth":

				config.SetAuthType("oauth")
				config.Log.Info("Authentication method set to OAuth")

			default:
				config.Log.Error("Invalid Selection")
			}

			return nil
		},
	}
	return cmd

}
