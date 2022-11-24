package parse

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/schollz/progressbar/v3"
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

func saveCanalLine(bytes []byte, dir string) {
	line := CANAL{}

	lineArray := strings.Split(string(bytes), " ")

	if len(lineArray) > 5 {
		line.Month = lineArray[0]
		line.Day = lineArray[1]
		line.Time = lineArray[2]
		line.HostName = lineArray[3]
		line.Service = strings.ReplaceAll(lineArray[4], ":", "")
		line.Message = strings.ReplaceAll(strings.Join(lineArray[5:], ""), "\n", "")

		if line.Month != "" && line.Day != "" && line.Time != "" && line.HostName != "" && line.Service != "" && line.Message != "" && line.HostName != "at" {
			folder := "/Standard/"
			if strings.Contains(line.Message, "Error") || strings.Contains(line.Message, "WARNING") {
				folder = "/Errors/"
			}
			filename := dir + line.HostName + "/" + line.Month + "-" + line.Day + folder + "/Canal.log"
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
			fileWriter := bufio.NewWriter(f)
			_, writeErr := fileWriter.WriteString(string(bytes))
			if writeErr != nil {
				panic(writeErr)
			}
			fileWriter.Flush()
			f.Close()
		}

	}

}

func newCanalCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "canal",
		Short:   "parse a canal log file",
		Long:    "Parse a canal log file.",
		Example: "sail parse canal /path/to/log.text | /path/to/log.log",
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

				fmt.Printf("Parsing %s, Output will be in %s\n", base, dir)

				bar := progressbar.DefaultBytes(fileinfo.Size(), "Parsing Canal")
				barWriter := io.Writer(bar)

				reader := bufio.NewReader(file)

				for {
					lineCount++
					token, err := reader.ReadBytes('\n')
					barWriter.Write(token)
					go saveCanalLine(token, dir)
					if err != nil {
						break
					}
				}

				fmt.Println("Finished Processing " + fmt.Sprint(lineCount) + " Lines")

			} else {
				return fmt.Errorf("please provide a filepath to the CANAL log file you wish to parse")
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
