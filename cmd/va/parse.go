package va

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

func ParseCCGFile(p *mpb.Progress, filepath string, everything bool) error {
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

	var wg sync.WaitGroup

	for {
		lineCount++
		token, err := bufReader.ReadBytes('\n')
		if err != nil {
			break
		} else {
			if ErrorCheck(token) || everything {
				var line CCG
				unErr := json.Unmarshal(token, &line)
				if unErr == nil {
					if line.Org != "" {
						processCount++
						wg.Add(1)
						go func() {
							defer wg.Done()
							if ErrorCheck(token) {
								saveCCGLine(line, dir, true)
							} else {
								saveCCGLine(line, dir, false)
							}
						}()
					}
				}
			}
		}
	}
	wg.Wait()
	bar.SetTotal(-1, true)

	return nil

}

func ParseCanalFile(p *mpb.Progress, filepath string, everything bool) error {
	var lineCount int

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

	var wg sync.WaitGroup

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

	return nil
}

func newParseCmd(client client.Client) *cobra.Command {
	var ccg bool
	var canal bool
	var everything bool
	cmd := &cobra.Command{
		Use:     "parse",
		Short:   "parse log files from a va",
		Long:    "Parse log files from a Virtual Appliance.",
		Example: "sail va parse ./path/to/ccg.log ./path/to/ccg.log ./path/to/canal.log ./path/to/canal.log",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ccg || canal {
				var wg sync.WaitGroup

				p := mpb.New(
					mpb.PopCompletedMode(),
					mpb.WithWidth(60),
					mpb.WithRefreshRate(180*time.Millisecond),
				)

				log.SetOutput(p)
				color.Blue("Parsing Files %s\n", args)
				for i := 0; i < len(args); i++ {
					wg.Add(1)

					filepath := args[i]

					if ccg {
						go func() {
							defer wg.Done()
							err := ParseCCGFile(p, filepath, everything)
							if err != nil {
								log.Panicf("Issue Parsing log file: %v", filepath)
							}
						}()
					} else if canal {
						go func() {
							defer wg.Done()
							err := ParseCanalFile(p, filepath, everything)
							if err != nil {
								log.Panicf("Issue Parsing log file: %v", filepath)
							}
						}()
					}
				}
				wg.Wait()

				return nil
			} else {
				return errors.New("must specify either ccg or canal")
			}
		},
	}

	cmd.Flags().BoolVarP(&ccg, "ccg", "", false, "Specifies the provided files are CCG Files")
	cmd.Flags().BoolVarP(&canal, "canal", "", false, "Specifies the provided files are CANAL Files")
	cmd.Flags().BoolVarP(&everything, "everything", "e", false, "Specifies all log traffic should be parsed, not just errors")
	cmd.MarkFlagsMutuallyExclusive("ccg", "canal")
	cmd.MarkFlagsMutuallyExclusive("everything", "canal")

	return cmd
}
