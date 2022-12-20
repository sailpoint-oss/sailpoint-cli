package va

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

func newTroubleshootCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "troubleshoot",
		Short:   "troubleshoot a va",
		Long:    "Troubleshoot a Virtual Appliance.",
		Example: "sail va troubleshoot 10.10.10.10",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output := cmd.Flags().Lookup("output").Value.String()
			header := "\n================== %v ==================\n"
			if output == "" {
				output, _ = os.Getwd()
			}
			var credentials []string
			for credential := 0; credential < len(args); credential++ {
				fmt.Printf("Enter Password for %v:", args[credential])
				password, _ := password()
				credentials = append(credentials, password)
			}

			for host := 0; host < len(args); host++ {
				endpoint := args[host]
				outputDir := path.Join(output, endpoint)
				outputFile := path.Join(outputDir, "Troubleshooting Log.log")

				if _, err := os.Stat(outputDir); errors.Is(err, os.ErrNotExist) {
					err := os.MkdirAll(outputDir, 0700)
					if err != nil {
						return err
					}
				}

				// Create the log file
				logFile, err := os.Create(outputFile)
				if err != nil {
					return err
				}
				defer logFile.Close()

				password := credentials[host]

				orgname, orgErr := runVACmd(endpoint, password, "awk '/org/' /home/sailpoint/config.yaml | sed 's/org: //' | sed 's/\r$//'")
				if orgErr != nil {
					return orgErr
				}
				orgname = strings.ReplaceAll(orgname, "\n", "")

				podname, podErr := runVACmd(endpoint, password, "awk '/pod/' /home/sailpoint/config.yaml | sed 's/pod: //' | sed 's/\r$//'")
				if podErr != nil {
					return podErr
				}
				podname = strings.ReplaceAll(podname, "\n", "")

				fmt.Printf("\n\nTroubleshooting VA\nHost: %v\nOrg: %v\nPod: %v\n\n", endpoint, orgname, podname)
				logFile.WriteString(fmt.Sprintf("Troubleshooting VA\nHost: %v\nOrg: %v\nPod: %v\n", endpoint, orgname, podname))

				config, configErr := runVACmd(endpoint, password, `cat config.yaml | sed "s/keyPassphrase:.*/keyPassphrase: <redacted>/g"`)
				if configErr != nil {
					return configErr
				}
				logFile.WriteString(fmt.Sprintf(header, "config.yaml"))
				logFile.WriteString(config)

				canal := strings.Contains(config, "tunnelTraffic: true")
				fmt.Printf("Secure Tunnel Configured: %v\n", canal)
				logFile.WriteString(fmt.Sprintf("Secure Tunnel Configured: %v\n", canal))
				if canal {
					color.Yellow("Found that Secure Tunnel (canal) service is enabled; additional tests will be run")
				}

				jdkversion, jdkErr := runVACmd(endpoint, password, "grep -a openjdk /home/sailpoint/log/worker.log | tail -1")
				if jdkErr != nil {
					return jdkErr
				}
				fmt.Printf("OpenJDK Version: %v", strings.ReplaceAll(jdkversion, "openjdk version ", ""))
				logFile.WriteString(fmt.Sprintf(header, "JDK Version"))
				logFile.WriteString(fmt.Sprintf("OpenJDK Version: %v", strings.ReplaceAll(jdkversion, "openjdk version ", "")))

				uname, unameErr := runVACmd(endpoint, password, "uname -a")
				if unameErr != nil {
					return unameErr
				}
				logFile.WriteString(fmt.Sprintf(header, "uname"))
				logFile.WriteString(uname)

				if strings.Contains(uname, "flatcar") {
					color.Green("OS Version: flatcar")
				} else {
					color.Red("OS Version: CoreOS")
				}

				profile, _ := runVACmd(endpoint, password, "cat /etc/profile.env")
				if profile != "" {
					color.Yellow("profile.env is present. May need to remove it if proxying is an issue.")
				} else {
					color.Green("profile.env is not present")
				}
				logFile.WriteString(fmt.Sprintf(header, "profile.env"))
				logFile.WriteString(profile)

				defaultEnv, _ := runVACmd(endpoint, password, "cat /etc/systemd/system.conf.d/10-default-env.conf")
				if defaultEnv != "" {
					color.Yellow("10-default-env.conf is present. May need to remove it if proxying is an issue.")
				} else {
					color.Green("10-default-env.conf is not present")
				}
				logFile.WriteString(fmt.Sprintf(header, "10-default-env.conf"))
				logFile.WriteString(defaultEnv)

				fmt.Println()

				color.Blue("Collecting docker.env")
				dockerEnv, _ := runVACmd(endpoint, password, "cat /home/sailpoint/docker.env")
				logFile.WriteString(fmt.Sprintf(header, "docker.env"))
				logFile.WriteString(dockerEnv)

				color.Blue("Collecting static.network")
				staticNetwork, _ := runVACmd(endpoint, password, "cat /etc/systemd/network/static.network")
				logFile.WriteString(fmt.Sprintf(header, "static.network"))
				logFile.WriteString(staticNetwork)

				color.Blue("Collecting resolve.conf")
				resolveConf, _ := runVACmd(endpoint, password, "cat /etc/resolv.conf")
				logFile.WriteString(fmt.Sprintf(header, "resolve.conf"))
				logFile.WriteString(resolveConf)

				color.Blue("Collecting proxy.yaml")
				proxyYaml, _ := runVACmd(endpoint, password, "cat /home/sailpoint/proxy.yaml")
				logFile.WriteString(fmt.Sprintf(header, "proxy.yaml"))
				logFile.WriteString(proxyYaml)

				color.Blue("Collecting os-release")
				osRelease, _ := runVACmd(endpoint, password, "cat /etc/os-release")
				logFile.WriteString(fmt.Sprintf(header, "os-release"))
				logFile.WriteString(osRelease)

				color.Blue("Collecting CPU Info")
				cpuInfo, _ := runVACmd(endpoint, password, "lscpu")
				logFile.WriteString(fmt.Sprintf(header, "lscpu"))
				logFile.WriteString(cpuInfo)

				color.Blue("Collecting RAM Info")
				ramInfo, _ := runVACmd(endpoint, password, "free")
				logFile.WriteString(fmt.Sprintf(header, "free"))
				logFile.WriteString(ramInfo)

				color.Blue("Collecting Network Info")
				networkInfo, _ := runVACmd(endpoint, password, "ifconfig")
				logFile.WriteString(fmt.Sprintf(header, "ifconfig"))
				logFile.WriteString(networkInfo)

				fmt.Println()

				color.Blue("Running Network Checks")
				networkChecks, _ := runVACmd(endpoint, password, `grep -a "Networking check" /home/sailpoint/log/charon.log`)
				logFile.WriteString(fmt.Sprintf(header, "Network Checks"))
				logFile.WriteString(networkChecks)

				color.Blue("Running sudo systemctl disable esx_dhcp_bump")
				logFile.WriteString(fmt.Sprintf(header, "sudo systemctl disable esx_dhcp_bump"))
				dhcpBump, _ := runVACmd(endpoint, password, "sudo systemctl disable esx_dhcp_bump")
				logFile.WriteString(dhcpBump)

				color.Blue("Running SQS Test")
				sqsTest, _ := runVACmd(endpoint, password, "curl -i -vv https://sqs.us-east-1.amazonaws.com")
				logFile.WriteString(fmt.Sprintf(header, "SQS Test"))
				logFile.WriteString(sqsTest)

				color.Blue("Running Tenant Test")
				tenantTest, _ := runVACmd(endpoint, password, fmt.Sprintf(`curl -i "https://%v.identitynow.com"`, orgname))
				logFile.WriteString(fmt.Sprintf(header, "Tenant Test"))
				logFile.WriteString(tenantTest)

				color.Blue("Running API Test")
				apiTest, _ := runVACmd(endpoint, password, fmt.Sprintf(`curl -i "https://%v.api.identitynow.com"`, orgname))
				logFile.WriteString(fmt.Sprintf(header, "API Test"))
				logFile.WriteString(apiTest)

				color.Blue("Running Pod Test")
				podTest, _ := runVACmd(endpoint, password, fmt.Sprintf(`curl -i "https://%v.accessiq.sailpoint.com"`, podname))
				logFile.WriteString(fmt.Sprintf(header, "Pod Test"))
				logFile.WriteString(podTest)

				color.Blue("Running DynamoDB Test")
				dynamodbTest, _ := runVACmd(endpoint, password, "curl -i https://dynamodb.us-east-1.amazonaws.com")
				logFile.WriteString(fmt.Sprintf(header, "dynamodb Test"))
				logFile.WriteString(dynamodbTest)
			}
			return nil
		},
	}

	cmd.Flags().StringP("endpoint", "e", "", "The host to troubleshoot")
	cmd.Flags().StringP("output", "o", "", "The path to save the log file")

	return cmd
}
