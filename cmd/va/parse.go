package va

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
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

var cache = make(map[string]*os.File)
var cacheLock sync.Mutex

const numWorkers = 8

func saveCanalLine(bytes []byte, dir string) {
	line := CANAL{}

	lineArray := strings.Split(string(bytes), " ")

	if len(lineArray) > 5 {
		line.Month = lineArray[0]
		line.Day = lineArray[1]
		line.Time = lineArray[2]
		line.HostName = lineArray[3]
		line.Service = strings.ReplaceAll(lineArray[4], ":", "")
		line.Message = strings.ReplaceAll(strings.Join(lineArray[5:], " "), "\n", "")

		if line.HostName != "at" && line.Month != "" && line.Day != "" && line.Time != "" && line.HostName != "" && line.Service != "" && line.Message != "" {
			folder := "Standard"
			if strings.Contains(line.Message, "Error") || strings.Contains(line.Message, "WARNING") {
				folder = "Errors"
			}
			filename := path.Join(dir, line.HostName, line.Month+"-"+line.Day, folder, "Canal.log")
			tempdir, _ := path.Split(filename)

			cacheLock.Lock()
			defer cacheLock.Unlock()

			f, exists := cache[filename]
			if !exists {
				if _, err := os.Stat(tempdir); errors.Is(err, os.ErrNotExist) {
					err := os.MkdirAll(tempdir, 0700)
					if err != nil {
						log.Error(err)
					}
				}
				f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					panic(err)
				}
				cache[filename] = f
			}

			fileWriter := bufio.NewWriter(f)
			_, writeErr := fileWriter.WriteString(string(bytes))
			if writeErr != nil {
				panic(writeErr)
			}
			fileWriter.Flush()
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

	cacheLock.Lock()
	defer cacheLock.Unlock()

	f, exists := cache[filename]
	if !exists {
		tempdir, _ := path.Split(filename)
		if _, err := os.Stat(tempdir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(tempdir, 0700)
			if err != nil {
				log.Error(err)
			}
		}
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		cache[filename] = f
	}

	if _, writeErr := f.Write(jsonBytes); writeErr != nil {
		return writeErr
	}

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

func ParseCCGFile(p *mpb.Progress, filepath string, all bool) error {
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

	type task struct {
		line  CCG
		token []byte
		dir   string
	}

	taskChan := make(chan task)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for t := range taskChan {
				saveCCGLine(t.line, t.dir, ErrorCheck(t.token))
			}
		}()
	}

	for {
		token, err := bufReader.ReadBytes('\n')
		if err != nil {
			break
		} else {
			if ErrorCheck(token) || all {
				var line CCG
				unErr := json.Unmarshal(token, &line)
				if unErr == nil && line.Org != "" {
					taskChan <- task{
						line:  line,
						token: token,
						dir:   dir,
					}
				}
			}
		}
	}
	close(taskChan)
	wg.Wait()
	bar.SetTotal(-1, true)

	return nil
}

func ParseCanalFile(p *mpb.Progress, filepath string, all bool) error {
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

	type task struct {
		token []byte
		dir   string
	}

	taskChan := make(chan task)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for t := range taskChan {
				saveCanalLine(t.token, t.dir)
			}
		}()
	}

	for {
		token, err := bufReader.ReadBytes('\n')
		if err != nil {
			break
		} else {
			taskChan <- task{
				token: token,
				dir:   dir,
			}
		}
	}
	close(taskChan)
	wg.Wait()
	bar.SetTotal(-1, true)

	return nil
}

//go:embed parse.md
var parseHelp string

func newParseCommand() *cobra.Command {
	help := util.ParseHelp(parseHelp)
	var fileType string
	var all bool
	cmd := &cobra.Command{
		Use:     "parse",
		Short:   "Parse Log Files from SailPoint Virtual Appliances",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			if fileType != "" {

				var wg sync.WaitGroup

				p := mpb.New(
					mpb.PopCompletedMode(),
					mpb.WithWidth(60),
					mpb.WithRefreshRate(180*time.Millisecond),
				)

				log.Info("Parsing Log Files", "files", args)

				log.SetOutput(p)
				for _, filepath := range args {
					wg.Add(1)

					switch fileType {
					case "ccg":
						go func(filepath string) {
							defer wg.Done()
							err := ParseCCGFile(p, filepath, all)
							if err != nil {
								log.Error("Issue Parsing log file", "file", filepath, "error", err)
							}
						}(filepath)
					case "canal":
						go func(filepath string) {
							defer wg.Done()
							err := ParseCanalFile(p, filepath, all)
							if err != nil {
								log.Error("Issue Parsing log file", "file", filepath, "error", err)
							}
						}(filepath)
					}

				}

				wg.Wait()
			} else {
				cmd.Help()
			}
			return nil

		},
	}

	cmd.Flags().StringVarP(&fileType, "type", "t", "", "Specifies the log type to parse (ccg, canal)")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Specifies all log traffic should be parsed, not just errors")

	return cmd
}
