package va

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/va"
	"github.com/spf13/cobra"
)

func updateAndRebootVA(endpoint, password string) {
	log.Info("Attempting to Update", "VA", endpoint)

	update, updateErr := va.RunVACmd(endpoint, password, UpdateCommand)
	if updateErr != nil {
		log.Error("Problem Updating", "VA", endpoint, "err", updateErr, "resp", update)
	} else {
		log.Info("Virtual Appliance Updating", "VA", endpoint)
		reboot, rebootErr := va.RunVACmd(endpoint, password, RebootCommand)
		if rebootErr != nil && rebootErr.Error() != "wait: remote command exited without exit status or exit signal" {
			log.Error("Problem Rebooting", "Server", endpoint, "err", rebootErr, "resp", reboot)
		} else {
			log.Info("Virtual Appliance Rebooting", "VA", endpoint)
		}
	}

	fmt.Println()
}

func newUpdateCommand(term terminal.Terminal) *cobra.Command {
	var credentials []string
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Perform Update Operations on a SailPoint Virtual Appliance",
		Long:    "\nPerform Update Operations on a SailPoint Virtual Appliance\n\n",
		Example: "sail va update 10.10.10.10 10.10.10.11 -c S@ilp0int -c S@ilp0int",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for i, endpoint := range args {

				password := credentials[i]

				if password == "" {
					password, _ = term.PromptPassword("Enter Password for " + endpoint + ":")
				}

				updateAndRebootVA(endpoint, password)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&credentials, "Passwords", "p", []string{}, "You can enter the Passwords for the servers in the same order that the servers are listed as arguments")

	return cmd
}
