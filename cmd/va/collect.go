package va

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/sftp"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/crypto/ssh"
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

			var wg sync.WaitGroup
			// passed wg will be accounted at p.Wait() call
			p := mpb.New(mpb.WithWidth(60),
				mpb.PopCompletedMode(),
				mpb.WithRefreshRate(180*time.Millisecond),
				mpb.WithWaitGroup(&wg))

			log.SetOutput(p)

			for credential := 0; credential < len(args); credential++ {
				password, _ := util.PromptPassword(fmt.Sprintf("Enter Password for %v:", args[credential]))
				credentials = append(credentials, password)
			}

			for host := 0; host < len(args); host++ {
				endpoint := args[host]
				log.Printf("Starting Collection for %v\n", endpoint)
				config := &ssh.ClientConfig{
					User:            "sailpoint",
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
					Auth: []ssh.AuthMethod{
						ssh.Password(credentials[host]),
					},
				}

				outputFolder := path.Join(output, endpoint)

				for file := 0; file < len(files); file++ {
					wg.Add(1)
					go func(file int) {
						// Connect
						client, err := ssh.Dial("tcp", net.JoinHostPort(endpoint, "22"), config)
						if err != nil {
							fmt.Println(err)
						}

						sftp, err := sftp.NewClient(client)
						if err != nil {
							fmt.Println(err)
						}
						defer sftp.Close()

						defer wg.Done()

						_, base := path.Split(files[file])
						outputFile := path.Join(outputFolder, base)
						if _, err := os.Stat(outputFolder); errors.Is(err, os.ErrNotExist) {
							err := os.MkdirAll(outputFolder, 0700)
							if err != nil {
								fmt.Println(err)
							}
						}
						remoteFileStats, statErr := sftp.Stat(files[file])
						if statErr == nil {
							name := fmt.Sprintf("%v - %v", endpoint, base)
							bar := p.AddBar(remoteFileStats.Size(),
								mpb.BarFillerClearOnComplete(),
								mpb.PrependDecorators(
									// simple name decorator
									decor.Name(name, decor.WCSyncSpaceR),
									decor.Name(":", decor.WCSyncSpaceR),
									decor.OnComplete(decor.CountersKiloByte("% .2f / % .2f", decor.WCSyncSpaceR), "Complete"),
									decor.TotalKiloByte("% .2f", decor.WCSyncSpaceR),
								),
								mpb.AppendDecorators(
									decor.OnComplete(decor.EwmaSpeed(decor.UnitKB, "% .2f", 90, decor.WCSyncWidth), ""),
									decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
								),
							)

							// Open the source file
							srcFile, err := sftp.Open(files[file])
							if err != nil {
								log.Println(err)
							}
							defer srcFile.Close()

							// Create the destination file
							dstFile, err := os.Create(outputFile)
							if err != nil {
								fmt.Println(err)
							}
							defer dstFile.Close()

							writer := io.Writer(dstFile)

							// create proxy reader
							proxyWriter := bar.ProxyWriter(writer)
							defer proxyWriter.Close()

							io.Copy(proxyWriter, srcFile)
						}
					}(file)
				}

			}
			p.Wait()
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
