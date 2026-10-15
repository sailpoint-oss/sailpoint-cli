package set

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
)

func PromptAuth() (string, error) {
	items := []tui.Choice{
		{Title: "PAT", Description: "Person Access Token - Single User PAT Configuration"},
		{Title: "OAuth", Description: "OAuth2.0 Authentication - Sign in via the Identity Security Cloud web portal"},
		{Title: "Bearer", Description: "Bearer Token - Use an access token already minted at the OAuth token endpoint"},
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
		Short:   "Set the currently active authentication mode (PAT, OAuth, bearer)",
		Long:    "\nSet the currently active authentication mode\n\nSupported Authentication Methods:\nPAT\nOAuth\nBearer - an access token already minted at the Identity Security Cloud OAuth token endpoint",
		Example: "sail set auth pat | sail set auth oauth | sail set auth bearer",
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
				log.Info("Authentication method set to PAT")

			case "oauth":

				config.SetAuthType("oauth")
				log.Info("Authentication method set to OAuth")

			case "bearer":

				config.SetAuthType("bearer")
				log.Info("Authentication method set to bearer")
				log.Info("Provide the minted access token with `sail set bearer`, SAIL_ACCESS_TOKEN, or --access-token")

			default:
				log.Warn("Unsupported authentication method. Choose pat, oauth, or bearer", "selection", selection)
			}

			return nil
		},
	}
	return cmd

}
