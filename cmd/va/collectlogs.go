package va

import (
	"fmt"
	"os"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

func newCollectLogsCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collect",
		Short:   "collect logs from a va",
		Long:    "Collect all relevant logs from a Virtual Appliance.",
		Example: "sail va collect -e 10.10.10.10",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("endpoint").Value.String()
			output := cmd.Flags().Lookup("output").Value.String()
			if endpoint != "" {
				if output == "" {
					output, _ = os.Getwd()
				}
				password, _ := password()
				ccgErr := getVAFile(endpoint, password, "/home/sailpoint/log/ccg.log", output)
				if ccgErr != nil {
					return ccgErr
				}
				charonErr := getVAFile(endpoint, password, "/home/sailpoint/log/charon.log", output)
				if charonErr != nil {
					return charonErr
				}
			} else {
				fmt.Println("an endpoint must be specified")
			}

			return nil
		},
	}

	cmd.Flags().StringP("endpoint", "e", "", "The host to collect logs from")
	cmd.Flags().StringP("output", "o", "", "The path to save the log files")

	return cmd
}
