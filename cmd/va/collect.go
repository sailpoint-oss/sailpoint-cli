package va

import (
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/va"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
)

func newCollectCmd(term terminal.Terminal) *cobra.Command {
	var credentials []string
	var output string
	var logs bool
	var config bool
	cmd := &cobra.Command{
		Use:     "collect",
		Short:   "Collect Configuration or Log Files from a SailPoint Virtual Appliance",
		Long:    "\nCollect Configuration or Log Files from a SailPoint Virtual Appliance\n\n",
		Example: "sail va collect 10.10.10.25, 10.10.10.26 -p S@ilp0int -p S@ilp0int \n\nLog Files:\n/home/sailpoint/log/ccg.log\n/home/sailpoint/log/charon.log\n/home/sailpoint/stuntlog.txt\n\nConfig Files:\n/home/sailpoint/proxy.yaml\n/etc/systemd/network/static.network\n/etc/resolv.conf\n",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

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

			var wg sync.WaitGroup
			p := mpb.New(mpb.WithWidth(60),
				mpb.PopCompletedMode(),
				mpb.WithRefreshRate(180*time.Millisecond),
				mpb.WithWaitGroup(&wg))

			for i, endpoint := range args {
				var password string

				if len(credentials) >= i-1 {
					password = credentials[i]
				}

				if password == "" {
					password, err = term.PromptPassword("Please enter the password for " + endpoint)
					if err != nil {
						return err
					}
				}
				wg.Add(1)
				go func(endpoint, password string) {
					defer wg.Done()
					outputFolder := output

					err := va.CollectVAFiles(endpoint, password, outputFolder, files, p)
					if err != nil {
						log.Error("Error collecting files for", "VA", endpoint, "err", err)
					}
				}(endpoint, password)
			}
			p.Wait()

			log.Info("All Operations Complete")
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "Output", "o", "", "The path to save the log files")
	cmd.Flags().BoolVarP(&logs, "logs", "l", false, "Retrieve log files")
	cmd.Flags().BoolVarP(&config, "config", "c", false, "Retrieve config files")
	cmd.Flags().StringArrayVarP(&credentials, "Passwords", "p", []string{}, "You can enter the Passwords for the servers in the same order that the servers are listed as arguments")

	cmd.MarkFlagsMutuallyExclusive("config", "logs")

	return cmd
}
