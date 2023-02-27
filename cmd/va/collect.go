package va

import (
	"fmt"
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/va"
	"github.com/spf13/cobra"
)

func newCollectCmd() *cobra.Command {
	var output string
	var logs bool
	var config bool
	cmd := &cobra.Command{
		Use:     "collect",
		Short:   "collect files from a virtual appliance",
		Long:    "Collect files from a Virtual Appliance.",
		Example: "sail va collect 10.10.10.10, 10.10.10.11 (-l only collect log files) (-c only collect config files) (-o /path/to/save/files)\n\nLog Files:\n/home/sailpoint/log/ccg.log\n/home/sailpoint/log/charon.log\n/home/sailpoint/stuntlog.txt\n\nConfig Files:\n/home/sailpoint/proxy.yaml\n/etc/systemd/network/static.network\n/etc/resolv.conf\n",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var credentials []string

			if output == "" {
				output, _ = os.Getwd()
			}
			var files []string
			if logs {
				files = []string{"/home/sailpoint/log/ccg.log", "/home/sailpoint/log/charon.log", "/home/sailpoint/stuntlog.txt"}
			} else if config {
				files = []string{"/home/sailpoint/proxy.yaml", "/etc/systemd/network/static.network", "/etc/resolv.conf"}
			} else {
				files = []string{"/home/sailpoint/log/ccg.log", "/home/sailpoint/log/charon.log", "/home/sailpoint/stuntlog.txt", "/home/sailpoint/proxy.yaml", "/etc/systemd/network/static.network", "/etc/resolv.conf"}
			}

			for credential := 0; credential < len(args); credential++ {
				password, _ := terminal.PromptPassword(fmt.Sprintf("Enter Password for %v:", args[credential]))
				credentials = append(credentials, password)
			}

			for host := 0; host < len(args); host++ {
				endpoint := args[host]
				password := credentials[host]
				outputFolder := path.Join(output, endpoint)

				err := va.CollectVAFiles(endpoint, password, outputFolder, files)
				if err != nil {
					return err
				}

			}
			color.Green("All Operations Complete")

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "Output", "o", "", "The path to save the log files")
	cmd.Flags().BoolVarP(&logs, "logs", "l", false, "Retrieve log files")
	cmd.Flags().BoolVarP(&config, "config", "c", false, "Retrieve config files")
	cmd.MarkFlagsMutuallyExclusive("config", "logs")

	return cmd
}
