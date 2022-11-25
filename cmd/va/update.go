package va

import (
	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

func newUpdateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "update a va",
		Long:    "update a Virtual Appliance.",
		Example: "sail va update -e 10.10.10.10",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("endpoint").Value.String()
			if endpoint != "" {
				password, _ := password()
				_, updateErr := runVACmd(endpoint, password, "sudo update_engine_client -check_for_update")
				if updateErr != nil {
					return updateErr
				} else {
					color.Green("Initiating update check and install (%v)", endpoint)
				}
				_, rebootErr := runVACmd(endpoint, password, "sudo reboot")
				if rebootErr != nil {
					color.Green("Rebooting Virtual Appliance (%v)", endpoint)
				} else {
					color.Red("Reboot failed")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("endpoint", "e", "", "The host to troubleshoot")
	cmd.Flags().StringP("output", "o", "", "The path to save the log file")

	return cmd
}
