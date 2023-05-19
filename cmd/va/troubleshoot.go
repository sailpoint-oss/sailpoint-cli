package va

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"os"
	"path"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/va"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
)

func NewTroubleshootCmd(term terminal.Terminal) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:     "troubleshoot",
		Short:   "Perform Troubleshooting Operations against a Virtual Appliance",
		Long:    "\nPerform Troubleshooting Operations against a Virtual Appliance\n\n",
		Example: "sail va troubleshoot 10.10.10.10",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if output == "" {
				output, _ = os.Getwd()
			}

			var credentials []string
			for credential := 0; credential < len(args); credential++ {
				password, _ := term.PromptPassword(fmt.Sprintf("Enter Password for %v:", args[credential]))
				credentials = append(credentials, password)
			}

			for host := 0; host < len(args); host++ {
				endpoint := args[host]
				outputDir := path.Join(output, endpoint)

				if _, err := os.Stat(outputDir); errors.Is(err, os.ErrNotExist) {
					err := os.MkdirAll(outputDir, 0700)
					if err != nil {
						return err
					}
				}

				password := credentials[host]

				orgErr := va.RunVACmdLive(endpoint, password, TroubleshootingScript)
				if orgErr != nil {
					return orgErr
				}

				color.Green("Troubleshooting Complete")
				color.Blue("Collecting stuntlog")

				var wg sync.WaitGroup
				p := mpb.New(mpb.WithWidth(60),
					mpb.PopCompletedMode(),
					mpb.WithRefreshRate(180*time.Millisecond),
					mpb.WithWaitGroup(&wg))

				err := va.CollectVAFiles(endpoint, password, outputDir, []string{"/home/sailpoint/stuntlog.txt"}, p)
				if err != nil {
					return err
				}

				color.Green("stuntlog copied to %s", outputDir)
			}

			return nil

		}}

	cmd.Flags().StringP("endpoint", "e", "", "The host to troubleshoot")
	cmd.Flags().StringVarP(&output, "output", "o", "", "The path to save the log file")

	return cmd

}
