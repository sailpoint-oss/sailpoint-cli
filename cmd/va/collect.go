package va

import (
	"os"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

func newCollectCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collect",
		Short:   "collect files from a va",
		Long:    "Collect files from a Virtual Appliance.",
		Example: "sail va collect -e 10.10.10.10 (-l log files) (-c config files) (-a all files)  (-o /path/to/save/files)",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("endpoint").Value.String()
			output := cmd.Flags().Lookup("output").Value.String()
			collectLogs := cmd.Flags().Lookup("logs").Value.String()
			collectConfig := cmd.Flags().Lookup("config").Value.String()
			collectAll := cmd.Flags().Lookup("all").Value.String()

			if endpoint != "" {
				if output == "" {
					output, _ = os.Getwd()
				}
				password, _ := password()

				if collectLogs == "true" || collectAll == "true" {
					ccgErr := getVAFile(endpoint, password, "/home/sailpoint/log/ccg.log", output)
					if ccgErr != nil {
						return ccgErr
					}

					charonErr := getVAFile(endpoint, password, "/home/sailpoint/log/charon.log", output)
					if charonErr != nil {
						return charonErr
					}

					stuntErr := getVAFile(endpoint, password, "/home/sailpoint/stuntlog.txt", output)
					if stuntErr != nil {
						color.Yellow("stuntlog.txt not found")
					}
				}
				if collectConfig == "true" || collectAll == "true" {
					configErr := getVAFile(endpoint, password, "/home/sailpoint/config.yaml", output)
					if configErr != nil {
						return configErr
					}

					proxyErr := getVAFile(endpoint, password, "/home/sailpoint/proxy.yaml", output)
					if proxyErr != nil {
						color.Yellow("proxy.yaml not found")
					}

					staticErr := getVAFile(endpoint, password, "/etc/systemd/network/static.network", output)
					if staticErr != nil {
						return staticErr
					}

					resolveErr := getVAFile(endpoint, password, "/etc/resolv.conf", output)
					if resolveErr != nil {
						return resolveErr
					}
				}

			} else {
				color.Red("an endpoint must be specified")
			}

			return nil
		},
	}

	cmd.Flags().StringP("endpoint", "e", "", "The host to collect logs from")
	cmd.Flags().StringP("output", "o", "", "The path to save the log files")
	cmd.Flags().BoolP("logs", "l", false, "Retrieve log files")
	cmd.Flags().BoolP("config", "c", false, "Retrieve config files")
	cmd.Flags().BoolP("all", "a", false, "Retrieve config files")

	return cmd
}
