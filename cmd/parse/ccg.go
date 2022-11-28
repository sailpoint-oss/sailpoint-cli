package parse

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

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

func CreateFolder(filepath string) error {
	tempdir, _ := path.Split(filepath)
	if _, err := os.Stat(tempdir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(tempdir, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveCCGLine(line CCG, dir string, isErr bool) error {
	folder := "Standard"
	if isErr {
		folder = "Errors"
	}
	filename := path.Join(dir, line.Org, folder, line.Timestamp.Format("2006-01-02"), strings.ReplaceAll(line.Logger_name, ".", "-"), "log.json")
	jsonBytes, _ := json.MarshalIndent(line, "", " ")
	err := CreateFolder(filename)
	if err != nil {
		return err
	}
	f, openErr := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		return openErr
	}
	if _, writeErr := f.Write(jsonBytes); writeErr != nil {
		return writeErr
	}
	f.Close()
	return nil
}

func ParseJSON(str string) []byte {
	var js json.RawMessage
	json.Unmarshal([]byte(str), &js)
	if js != nil {
		return js
	}
	return nil
}

func ErrorCheck(token []byte) bool {
	errorString := []byte("error")
	exceptionString := []byte("exception")
	return bytes.Contains(token, errorString) || bytes.Contains(token, exceptionString)
}

func newCCGCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ccg",
		Short:   "parse a CCG log file",
		Long:    "Parse a CCG log file.\n\nBy default, only errors are parsed out\nTo parse everything use -e",
		Example: "sail parse ccg /path/to/log.text | /path/to/log.log (-e parse all traffic)",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			everything := cmd.Flags().Lookup("everything").Value.String()

			var wg sync.WaitGroup

			p := mpb.New(
				mpb.PopCompletedMode(),
				mpb.WithWidth(60),
				mpb.WithRefreshRate(180*time.Millisecond),
			)

			log.SetOutput(p)

			fmt.Printf("Parsing %s\n", args)
			for i := 0; i < len(args); i++ {
				wg.Add(1)

				go func(i int) error {
					defer wg.Done()
					filepath := args[i]
					var lineCount int
					var processCount int

					file, err := os.Open(filepath)
					if err != nil {
						return err
					}
					defer file.Close()
					fileinfo, err := os.Stat(filepath)
					if err != nil {
						return err
					}

					dir, base := path.Split(filepath)

					bar := p.AddBar(fileinfo.Size(),
						mpb.PrependDecorators(
							decor.Name(fmt.Sprintf("%v", base), decor.WCSyncSpaceR),
							decor.CountersKiloByte("% .2f / % .2f", decor.WCSyncSpaceR),
						),
						mpb.AppendDecorators(
							decor.Name("["),
							decor.Percentage(),
							decor.Name("]["),
							decor.Elapsed(decor.ET_STYLE_GO),
							decor.Name(" Elapsed]"),
						),
					)

					proxyReader := bar.ProxyReader(file)
					defer proxyReader.Close()

					bufReader := bufio.NewReader(proxyReader)

					var iwg sync.WaitGroup

					for {
						lineCount++
						token, err := bufReader.ReadBytes('\n')
						if err != nil {
							break
						} else {
							if ErrorCheck(token) || everything == "true" {
								var line CCG
								unErr := json.Unmarshal(token, &line)
								if unErr == nil {
									if line.Org != "" {
										processCount++
										iwg.Add(1)
										go func() {
											defer iwg.Done()
											if ErrorCheck(token) {
												saveCCGLine(line, dir, true)
											} else {
												saveCCGLine(line, dir, false)
											}
										}()
										// if date == "" {
										// 	date = line.Timestamp.Format("2006-01-02")
										// }
										// if date != line.Timestamp.Format("2006-01-02") {
										// wg.Add(1)
										// funcQueue, funcDate := queue, date
										// queue = []CCG{}
										// go func(funcQueue []CCG, funcDate string) {
										// 	queueCount = queueCount + len(funcQueue)
										// 	for i := 0; i < len(funcQueue)-1; i++ {
										// 		err := saveCCGLine(funcQueue[i], dir)
										// 		if err != nil {
										// 			log.Panic(err)
										// 		}
										// 	}
										// 	// defer log.Println("FinishedProcessing Queue (", funcDate, ") (", len(funcQueue), ")")
										// 	defer wg.Done()
										// }(funcQueue, funcDate)
										// date = line.Timestamp.Format("2006-01-02")
										// }
										// queue = append(queue, line)
										// filename := dir + line.Org + "/" + line.Timestamp.Format("2006-01-02") + folder + strings.ReplaceAll(line.Logger_name, ".", "-") + "/log.json"
										// index := indexOfPath(filename, Files)
										// if index != -1 {
										// 	Files[index].CCGEntries = append(Files[index].CCGEntries, line)
										// } else {
										// 	Files = append(Files, OutputFile{[]CCG{line}, filename})
										// }
										// bar.IncrBy(1)
										// bar.SetTotal(int64(lineCount), false)
									}
								}
							}

							// 	folder := "/Standard/"
							// 	if strings.Contains(line.Message, "error") || strings.Contains(line.Message, "exception") {
							// 		folder = "/Errors/"
							// 	}

							// 	if line.Org != "" {
							// 		filename := dir + line.Org + "/" + line.Timestamp.Format("2006-01-02") + folder + strings.ReplaceAll(line.Logger_name, ".", "-") + "/log.json"
							// 		jsonBytes, _ := json.MarshalIndent(line, "", " ")

							// 		f, openErr := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
							// 		if openErr != nil {
							// 			panic(openErr)
							// 		}

							// 		if _, writeErr := f.Write(jsonBytes); writeErr != nil {
							// 			panic(writeErr)
							// 		}
							// 		defer f.Close()
							// 		bar.EwmaIncrBy(len(token), time.Since(start))
							// 	}

						}
					}
					iwg.Wait()
					bar.SetTotal(-1, true)
					// fmt.Println("Processed", processCount, "lines of", lineCount, "Total Lines")

					return nil
				}(i)
			}
			wg.Wait()

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")
	cmd.Flags().BoolP("everything", "e", false, "parse all logs contents")

	return cmd
}
