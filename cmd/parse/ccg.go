package parse

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/schollz/progressbar/v3"
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

func saveCCGLine(bytes []byte, dir string) {
	line := CCG{}
	json.Unmarshal(bytes, &line)
	folder := "/Standard/"
	if strings.Contains(line.Message, "error") || strings.Contains(line.Message, "exception") {
		folder = "/Errors/"
	}
	if line.Org != "" {
		filename := dir + line.Org + "/" + line.Timestamp.Format("2006-01-02") + folder + strings.ReplaceAll(line.Logger_name, ".", "-") + "/log.json"
		jsonBytes, _ := json.MarshalIndent(line, "", " ")
		tempdir, _ := path.Split(filename)
		if _, err := os.Stat(tempdir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(tempdir, 0700)
			if err != nil {
				log.Println(err)
			}
		}
		f, openErr := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if openErr != nil {
			panic(openErr)
		}

		if _, writeErr := f.Write(jsonBytes); writeErr != nil {
			panic(writeErr)
		}
		defer f.Close()
	}
}

func newCCGCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ccg",
		Short:   "parse a CCG log file",
		Long:    "Parse a CCG log file.",
		Example: "sail parse ccg /path/to/log.text | /path/to/log.log",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var lineCount int
			filepath := cmd.Flags().Lookup("file").Value.String()
			if filepath != "" {
				file, err := os.Open(filepath)
				if err != nil {
					return err
				}
				defer file.Close()
				fileinfo, err := os.Stat(filepath)
				if err != nil {
					return err
				}
				fmt.Printf("Name:  %+v\nBytes: %+v\n", fileinfo.Name(), fileinfo.Size())

				dir, base := path.Split(filepath)

				fmt.Printf("Parsing %s\nOutput will be in %s\n", base, dir)

				bar := progressbar.DefaultBytes(fileinfo.Size(), "Parsing CCG")
				barWriter := io.Writer(bar)

				reader := bufio.NewReader(file)

				for {
					lineCount++
					token, err := reader.ReadBytes('\n')
					barWriter.Write(token)
					go saveCCGLine(token, dir)
					if err != nil {
						break
					}
				}

				fmt.Println("Finished Processing " + fmt.Sprint(lineCount) + " Lines")
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
