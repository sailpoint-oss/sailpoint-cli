package va

import (
	_ "embed"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/sailpoint-oss/sailpoint-cli/internal/va"
	"github.com/spf13/cobra"
)

//go:embed update.md
var updateHelp string

func updateAndRebootVA(endpoint, password string) {
	log.Info("Attempting to update", "VA", endpoint)

	update, updateErr := va.RunVACmd(endpoint, password, UpdateCommand)
	if updateErr != nil {
		log.Error("Problem updating", "VA", endpoint, "err", updateErr, "resp", update)
	} else {
		log.Info("Virtual appliance updating", "VA", endpoint)
		reboot, rebootErr := va.RunVACmd(endpoint, password, RebootCommand)
		if rebootErr != nil && rebootErr.Error() != "wait: remote command exited without exit status or exit signal" {
			log.Error("Problem rebooting", "Server", endpoint, "err", rebootErr, "resp", reboot)
		} else {
			log.Info("Virtual appliance rebooting", "VA", endpoint)
		}
	}

	fmt.Println()
}

func newUpdateCommand() *cobra.Command {
	help := util.ParseHelp(updateHelp)
	var credentials []string
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Perform update operations on a SailPoint virtual appliance",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("passwords") {
				log.Warn("Passing passwords as flags can expose them in shell history and process listings. Omit --passwords to use the secure prompt.")
			}
			for i, endpoint := range args {
				var password string

				if len(credentials) > i {
					password = credentials[i]
				}

				if password == "" {
					var err error
					password, err = tui.Password("Enter password for " + endpoint)
					if err != nil {
						return err
					}
				}

				updateAndRebootVA(endpoint, password)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&credentials, "passwords", "p", []string{}, "Passwords for the servers in the same order that the servers are listed as arguments")
	cmd.Flags().MarkDeprecated("passwords", "omit the flag to use the secure prompt")

	return cmd
}
