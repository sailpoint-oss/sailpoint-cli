package va

import (
	"fmt"
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
		Example: "sail va troubleshoot -e 10.10.10.10",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("endpoint").Value.String()
			if endpoint != "" {
				password, _ := password()
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
				fmt.Printf("\n\n\nTroubleshooting VA: %v\nOrg: %v\nPod: %v\n", endpoint, orgname, podname)
				config, configErr := runVACmd(endpoint, password, `cat config.yaml | sed "s/keyPassphrase:.*/keyPassphrase: <redacted>/g"`)
				if configErr != nil {
					return configErr
				}

				canal := strings.Contains(config, "tunnelTraffic: true")
				fmt.Printf("Secure Tunnel Configured: %v\n", canal)
				if canal {
					color.Yellow("Found that Secure Tunnel (canal) service is enabled; additional tests will be run")
				}

				uname, unameErr := runVACmd(endpoint, password, "uname -a")
				if unameErr != nil {
					return unameErr
				}

				if strings.Contains(uname, "flatcar") {
					color.Green("OS Version: flatcar")
				} else {
					color.Red("OS Version: CoreOS")
				}

				jdkversion, jdkErr := runVACmd(endpoint, password, "grep -a openjdk /home/sailpoint/log/worker.log | tail -1")
				if jdkErr != nil {
					return jdkErr
				}
				fmt.Printf("OpenJDK Version: %v", strings.ReplaceAll(jdkversion, "openjdk version ", ""))

				profile, _ := runVACmd(endpoint, password, "cat /etc/profile.env")
				if profile != "" {
					color.Yellow("profile.env is present. May need to remove it if proxying is an issue.")
				} else {
					color.Green("profile.env is not present")
				}

				defaultEnv, _ := runVACmd(endpoint, password, "cat /etc/systemd/system.conf.d/10-default-env.conf")
				if defaultEnv != "" {
					color.Yellow("10-default-env.conf is present. May need to remove it if proxying is an issue.")
				} else {
					color.Green("10-default-env.conf is not present")
				}

				dockerEnv, _ := runVACmd(endpoint, password, "cat /home/sailpoint/docker.env")
				fmt.Println(dockerEnv)

				staticNetwork, _ := runVACmd(endpoint, password, "cat /etc/systemd/network/static.network")
				fmt.Println(staticNetwork)

				resolveConf, _ := runVACmd(endpoint, password, "cat /etc/resolv.conf")
				fmt.Println(resolveConf)

				proxyYaml, _ := runVACmd(endpoint, password, "cat /home/sailpoint/proxy.yaml")
				fmt.Println(proxyYaml)

				osRelease, _ := runVACmd(endpoint, password, "cat /etc/os-release")
				fmt.Println(osRelease)

				cpuInfo, _ := runVACmd(endpoint, password, "lscpu")
				fmt.Println(cpuInfo)

				ramInfo, _ := runVACmd(endpoint, password, "free")
				fmt.Println(ramInfo)

				networkInfo, _ := runVACmd(endpoint, password, "ifconfig")
				fmt.Println(networkInfo)

				networkChecks, _ := runVACmd(endpoint, password, `grep -a "Networking check" /home/sailpoint/log/charon.log`)
				fmt.Println(networkChecks)

				dhcpBump, _ := runVACmd(endpoint, password, "sudo systemctl disable esx_dhcp_bump")
				fmt.Println(dhcpBump)

				sqsTest, _ := runVACmd(endpoint, password, "curl -i -vv https://sqs.us-east-1.amazonaws.com")
				fmt.Println(sqsTest)

				tenantTest, _ := runVACmd(endpoint, password, fmt.Sprintf(`curl -i "https://%v.identitynow.com"`, orgname))
				fmt.Println(tenantTest)

				apiTest, _ := runVACmd(endpoint, password, fmt.Sprintf(`curl -i "https://%v.api.identitynow.com"`, orgname))
				fmt.Println(apiTest)

				podTest, _ := runVACmd(endpoint, password, fmt.Sprintf(`curl -i "https://%v.accessiq.sailpoint.com"`, podname))
				fmt.Println(podTest)

				dynamodbTest, _ := runVACmd(endpoint, password, "curl -i https://dynamodb.us-east-1.amazonaws.com")
				fmt.Println(dynamodbTest)
			}

			return nil
		},
	}

	cmd.Flags().StringP("endpoint", "e", "", "The host to troubleshoot")

	return cmd
}
