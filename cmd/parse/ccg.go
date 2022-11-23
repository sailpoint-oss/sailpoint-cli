package parse

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func saveLine(line CCG, dir string) {

	// str := fmt.Sprintf("%#v", line)
	folder := "/Standard/"
	if strings.Contains(line.Message, "error") || strings.Contains(line.Message, "exception") {
		folder = "/Errors/"
	}
	if line.Org != "" {
		filename := dir + line.Org + "/" + line.Timestamp.Format("2006-01-02") + folder + strings.ReplaceAll(line.Logger_name, ".", "-") + "/log.json"
		bytes, _ := json.MarshalIndent(line, "", " ")
		tempdir, _ := path.Split(filename)
		if _, err := os.Stat(tempdir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(tempdir, 0700)
			if err != nil {
				log.Println(err)
			}
		}
		f, openErr := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		check(openErr)
		fileWriter := bufio.NewWriter(f)
		_, writeErr := fileWriter.Write(bytes)
		check(writeErr)
		// if _, err = f.Write(bytes); err != nil {
		// 	panic(err)
		// }
		f.Close()
	}
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
				barWriter := bufio.NewWriter(bar)
				if err := scanner.Err(); err != nil {
					return err
				}

				for scanner.Scan() {
					lineCount++
					bytes := scanner.Bytes()
					barWriter.Write(bytes)
					var line CCG
					json.Unmarshal(bytes, &line)
					go saveLine(line, dir)
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
