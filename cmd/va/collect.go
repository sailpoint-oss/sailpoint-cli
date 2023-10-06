package va

import (
	_ "embed"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/sailpoint-oss/sailpoint-cli/internal/va"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
)

//go:embed collect.md
var collectHelp string

func newCollectCommand(term terminal.Terminal) *cobra.Command {
	help := util.ParseHelp(collectHelp)
	var credentials []string
	var output string
	var logs bool
	var config bool
	cmd := &cobra.Command{
		Use:     "collect [-c | -l] [-o output] VA-Network-Address... [-p va-password]",
		Short:   "Collect files from a SailPoint Virtual Appliance",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			logFiles := []string{"/home/sailpoint/log/ccg.log", "/home/sailpoint/log/charon.log"}
			configFiles := []string{"/home/sailpoint/proxy.yaml", "/etc/systemd/network/static.network", "/etc/resolv.conf"}

			if output == "" {
				output, _ = os.Getwd()
			}
			var files []string
			if logs {
				files = append(files, logFiles...)
			}
			if config {
				files = append(files, configFiles...)
			}

			if !config && !logs {
				files = append(files, logFiles...)
				files = append(files, configFiles...)
			}

			var wg sync.WaitGroup
			p := mpb.New(
				mpb.PopCompletedMode(),
				mpb.WithRefreshRate(180*time.Millisecond),
				mpb.WithWaitGroup(&wg))

			log.SetOutput(p)

			for i, endpoint := range args {
				var password string

				if len(credentials) > i {
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

	cmd.Flags().StringVarP(&output, "output", "o", "", "The path to save the log files")
	cmd.Flags().BoolVarP(&logs, "log", "l", false, "retrieve log files")
	cmd.Flags().BoolVarP(&config, "config", "c", false, "retrieve config files")
	cmd.Flags().StringArrayVarP(&credentials, "passwords", "p", []string{}, "passwords for the servers in the same order that the servers are listed as arguments")

	cmd.MarkFlagsMutuallyExclusive("config", "log")

	return cmd
}
