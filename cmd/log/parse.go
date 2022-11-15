package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

type CCG struct {
	Stack                     string    `json:"stack"`
	Pod                       string    `json:"pod"`
	ConnectorLogging          string    `json:"connector-logging"`
	ClusterId                 string    `json:"clusterId"`
	BuildNumber               string    `json:"buildNumber"`
	ApiUsername               string    `json:"ApiUsername"`
	OrgType                   string    `json:"orgType"`
	File                      string    `json:"file"`
	Encryption                string    `json:"encryption"`
	ConnectorBundleIdentityiq string    `json:"connector-bundle-identityiq"`
	Line_number               int       `json:"line_number"`
	Version                   int       `json:"@version"`
	Logger_name               string    `json:"logger_name"`
	MantisClient              string    `json:"mantis-client"`
	Class                     string    `json:"class"`
	ClientId                  string    `json:"clientId"`
	Source_host               string    `json:"source_host"`
	Method                    string    `json:"method"`
	Org                       string    `json:"org"`
	Level                     string    `json:"level"`
	IdentityIQ                string    `json:"identityIQ"`
	Message                   string    `json:"message"`
	Pipeline                  string    `json:"pipeline"`
	Timestamp                 time.Time `json:"@timestamp"`
	Thread_name               string    `json:"thread_name"`
	Metrics                   string    `json:"metrics"`
	Region                    string    `json:"region"`
	Queue                     string    `json:"queue"`
	SCIMCommon                string    `json:"SCIM Common"`
}

func saveLine() {

}

func newParseCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "parse",
		Short:   "parse a log file",
		Long:    "Parse a log file Currently only supports CCG Logs.",
		Example: "sail log p /path/to/log.text | /path/to/log.log",
		Aliases: []string{"p"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var line CCG

			filepath := cmd.Flags().Lookup("file").Value.String()
			if filepath != "" {
				file, err := os.Open(filepath)
				if err != nil {
					return err
				}
				fileinfo, err := os.Stat(filepath)
				if err != nil {
					return err
				}
				fmt.Println(fileinfo)
				defer file.Close()

				dir, base := path.Split(filepath)

				fmt.Printf("Parsing %s, Output will be in %s\n", base, dir)

				scanner := bufio.NewScanner(file)
				if err := scanner.Err(); err != nil {
					return err
				}

				for scanner.Scan() {
					err := json.Unmarshal(scanner.Bytes(), &line)
					if err != nil {
						// fmt.Println(err)
						// fmt.Println(scanner.Text())
					}
					fmt.Printf("%+v\n", line)
				}
			} else {
				scanner := bufio.NewScanner(os.Stdin)
				if err := scanner.Err(); err != nil {
					return err
				}

				for scanner.Scan() {
					err := json.Unmarshal(scanner.Bytes(), &line)
					if err != nil {
						// fmt.Println(err)
						// fmt.Println(scanner.Text())
					}
					fmt.Printf("%+v\n", line)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
