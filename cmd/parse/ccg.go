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

func newCCGCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ccg",
		Short:   "parse a CCG log file",
		Long:    "Parse a CCG log file.",
		Example: "sail parse ccg /path/to/log.text | /path/to/log.log",
		Aliases: []string{"c"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var line CCG
			var lineCount int
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
				fmt.Printf("Name:  %+v\nBytes: %+v\n", fileinfo.Name(), fileinfo.Size())
				defer file.Close()

				dir, base := path.Split(filepath)

				fmt.Printf("Parsing %s, Output will be in %s\n", base, dir)
				bar := progressbar.DefaultBytes(fileinfo.Size(), "Parsing")
				scanner := bufio.NewScanner(file)
				barWriter := io.Writer(bar)
				if err := scanner.Err(); err != nil {
					return err
				}

				for scanner.Scan() {
					lineCount++
					err := json.Unmarshal(scanner.Bytes(), &line)
					if err != nil {
						// fmt.Println(err)
						// fmt.Println(scanner.Text())
					}
					// fmt.Printf("%+v\n", line)
					str := fmt.Sprintf("%#v", line)
					folder := "/Standard/"
					if strings.Contains(str, "error") || strings.Contains(str, "exception") {
						folder = "/Errors/"
					}
					if line.Org != "" {
						filename := dir + line.Org + "/" + line.Timestamp.Format("2006-01-02") + folder + strings.ReplaceAll(line.Logger_name, ".", "-") + "/log.json"
						tempdir, _ := path.Split(filename)
						if _, err := os.Stat(tempdir); errors.Is(err, os.ErrNotExist) {
							err := os.MkdirAll(tempdir, os.ModePerm)
							if err != nil {
								log.Println(err)
							}
						}
						f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						if err != nil {
							panic(err)
						}
						barWriter.Write(scanner.Bytes())
						if _, err = f.WriteString(strings.ReplaceAll(str, "log.CCG", "") + "\n"); err != nil {
							panic(err)
						}
						f.Close()
					}
				}
				fmt.Println("Finished Processing " + fmt.Sprint(lineCount) + " Lines")

			} else {
				return fmt.Errorf("please provide a filepath to the CCG log file you wish to parse")
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
