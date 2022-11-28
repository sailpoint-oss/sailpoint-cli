package parse

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
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
				color.Green("Name:  %+v\nBytes: %+v\n", fileinfo.Name(), fileinfo.Size())

				dir, base := path.Split(filepath)

				color.Green("Parsing %s\nOutput will be in %s\n", base, dir)

				var wg sync.WaitGroup
				p := mpb.New(
					mpb.WithWidth(60),
					mpb.WithRefreshRate(180*time.Millisecond),
				)

				bar := p.AddBar(fileinfo.Size(),
					mpb.PrependDecorators(
						decor.Name(fmt.Sprintf("%v | ", base)),
						decor.CountersKiloByte("% .2f / % .2f"),
					),
					mpb.AppendDecorators(
						decor.Name("["),
						decor.Percentage(),
						decor.Name("]"),
						decor.Name("["),
						decor.Elapsed(decor.ET_STYLE_GO),
						decor.Name(" Elapsed]"),
					),
				)

				proxyReader := bar.ProxyReader(file)
				defer proxyReader.Close()

				bufReader := bufio.NewReader(proxyReader)

				log.SetOutput(p)

				for {
					lineCount++
					token, err := bufReader.ReadBytes('\n')
					if err != nil {
						break
					} else {
						wg.Add(1)
						go func() {
							defer wg.Done()
							saveCanalLine(token, dir)
						}()
					}

				}
				wg.Wait()
				bar.SetTotal(-1, true)
				log.Println("Finished Processing " + fmt.Sprint(lineCount) + " Lines")

			} else {
				return fmt.Errorf("please provide a filepath to the CANAL log file you wish to parse")
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
