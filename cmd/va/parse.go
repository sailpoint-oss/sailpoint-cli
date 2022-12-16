package va

import (
	"time"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

type CANAL struct {
	Month    string
	Day      string
	Time     string
	HostName string
	Service  string
	Message  string
}

type CCG struct {
	Exception                 string    `json:"exception"`
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

func newParseCmd(client client.Client) *cobra.Command {
	var ccg bool
	var canal bool
	cmd := &cobra.Command{
		Use:     "parse",
		Short:   "parse log files from a va",
		Long:    "Parse log files from a Virtual Appliance.",
		Example: "sail va parse ./path/to/ccg.log ./path/to/ccg.log ./path/to/canal.log ./path/to/canal.log",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}

	cmd.Flags().BoolVarP(&ccg, "ccg", "ccg", false, "Specifies the provided files are CCG Files")
	cmd.Flags().BoolVarP(&canal, "canal", "canal", false, "Specifies the provided files are CANAL Files")
	cmd.MarkFlagsMutuallyExclusive("ccg", "canal")

	return cmd
}
