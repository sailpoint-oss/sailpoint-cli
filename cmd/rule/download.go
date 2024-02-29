// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package rule

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	SailPointHeader = `<!DOCTYPE Rule PUBLIC "sailpoint.dtd" "sailpoint.dtd">`
)

type Rule struct {
	XMLName     xml.Name  `xml:"Rule"`
	Name        string    `xml:"name,attr"`
	Type        string    `xml:"type,attr"`
	Description string    `xml:"Description"`
	Signature   Signature `xml:"Signature"`
	Source      string    `xml:"Source"`
}

type Signature struct {
	XMLName    xml.Name `xml:"Signature"`
	ReturnType string   `xml:"returnType,attr"`
	Inputs     Inputs   `xml:"Inputs"`
}

type Inputs struct {
	XMLName  xml.Name   `xml:"Inputs"`
	Argument []Argument `xml:"Argument"`
}

type Argument struct {
	XMLName     xml.Name `xml:"Argument"`
	Name        string   `xml:"name,attr"`
	Type        string   `xml:"type,attr"`
	Description string   `xml:"Description"`
}

type dict map[string]interface{}

func (d dict) d(k string) dict {
	return d[k].(map[string]interface{})
}

func (d dict) s(k string) string {
	return d[k].(string)
}

func newDownloadCommand() *cobra.Command {

	var description = string("Export of all rules for Download")
	var objectOptions string
	var includeTypes = []string{"RULE"}
	var excludeTypes []string

	return &cobra.Command{
		Use:     "list",
		Short:   "List all cloud rules in IdentityNow",
		Long:    "\nList all rules in IdentityNow\n\n",
		Example: "sail rule list | sail rule ls",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			var options *map[string]beta.ObjectExportImportOptions

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			if objectOptions != "" {
				err = json.Unmarshal([]byte(objectOptions), &options)
				if err != nil {
					return err
				}
			}

			job, _, err := apiClient.Beta.SPConfigAPI.ExportSpConfig(context.TODO()).ExportPayload(beta.ExportPayload{Description: &description, IncludeTypes: includeTypes, ExcludeTypes: excludeTypes, ObjectOptions: options}).Execute()
			if err != nil {
				return err
			}

			var entries [][]string

			time.Sleep(2 * time.Second)

			for {
				response, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExportStatus(context.TODO(), job.JobId).Execute()
				if err != nil {
					fmt.Println("Error YO")
					return err
				}
				if response.Status == "NOT_STARTED" || response.Status == "IN_PROGRESS" {
					color.Yellow("Status: %s. checking again in 5 seconds", response.Status)
					time.Sleep(5 * time.Second)
				} else {
					switch response.Status {
					case "COMPLETE":
						log.Info("Job Complete")
						exportData, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExport(context.TODO(), job.JobId).Execute()
						if err != nil {
							return err
						}

						// Save xml files
						for _, v := range exportData.Objects {
							//Make Rule XML Object
							rule := &Rule{Name: v.Object["name"].(string), Type: v.Object["type"].(string), Description: v.Object["description"].(string), Source: v.Object["sourceCode"]["script"].(string)}

							entries = append(entries, []string{v.Object["id"].(string), v.Object["name"].(string)})
						}

						output.WriteTable(cmd.OutOrStdout(), []string{"Id", "Name"}, entries)

						return nil
					case "CANCELLED":
						return fmt.Errorf("export task cancelled")
					case "FAILED":
						return fmt.Errorf("export task failed")
					}
					break
				}
			}

			return nil
		},
	}
}
