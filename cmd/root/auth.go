package root

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
)

func PromptAuth() (string, error) {
	items := []tui.Choice{
		{Title: "PAT", Description: "Person Access Token - Single User"},
		{Title: "OAuth", Description: "OAuth2.0 Authentication - Sign in via the Web Portal"},
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
		Long:    "Change Auth Mode Configured (pat, oauth).",
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
				if config.GetAuthType() != "pat" {
					config.SetAuthType("pat")
					color.Blue("authentication method set to pat")
				}
			case "oauth":
				if config.GetAuthType() != "oauth" {
					config.SetAuthType("oauth")
					color.Blue("authentication method set to oauth")
				}
			default:
				return fmt.Errorf("invalid selection")
			}

			return nil
		},
	}
	return cmd

}
