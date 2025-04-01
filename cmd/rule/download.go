// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package rule

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	SailPointHeader = `<!DOCTYPE Rule PUBLIC "sailpoint.dtd" "sailpoint.dtd">` + "\n"
)

type Rule struct {
	XMLName     xml.Name `xml:"Rule"`
	Name        string   `xml:"name,attr"`
	Type        string   `xml:"type,attr"`
	Description string   `xml:"Description,omitempty"`
	Signature   *Signature
	Source      string `xml:"Source"`
}

type Signature struct {
	XMLName    xml.Name `xml:"Signature"`
	ReturnType string   `xml:"returnType,attr"`
	Inputs     *Inputs
	Returns    *Returns
}

type Inputs struct {
	XMLName  xml.Name   `xml:"Inputs"`
	Argument []Argument `xml:"Argument"`
}

type Returns struct {
	XMLName  xml.Name   `xml:"Returns"`
	Argument []Argument `xml:"Argument"`
}

type Argument struct {
	XMLName     xml.Name `xml:"Argument"`
	Name        string   `xml:"name,attr"`
	Type        string   `xml:"type,attr,omitempty"`
	Description string   `xml:"Description,omitempty"`
}

var cloudRuleTypes = []string{"AttributeGenerator", "AttributeGeneratorFromTemplate", "BeforeProvisioning", "BuildMap", "Correlation", "IdentityAttribute", "ManagerCorrelation"}

func newDownloadCommand() *cobra.Command {

	var description = string("Export of all rules for Download")
	var objectOptions string
	var includeTypes = []string{"RULE"}
	var excludeTypes []string
	var destination string
	var cloud bool
	var connector bool

	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Download all rules in Identity Security Cloud",
		Long:    "\nDownload all rules in Identity Security Cloud\n\n",
		Example: "sail rule download",
		Aliases: []string{"d"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			var options *map[string]beta.ObjectExportImportOptions

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			if objectOptions != "" {
				err = json.Unmarshal([]byte(objectOptions), &options)
				if err != nil {
					return err
				}
			}

			if cloud {
				saveCloudXMLRules(apiClient, description, includeTypes, excludeTypes, options, destination)
			} else if connector {
				saveJSONConnectorRules(apiClient, destination)
			} else {
				saveCloudXMLRules(apiClient, description, includeTypes, excludeTypes, options, destination)
				saveJSONConnectorRules(apiClient, destination)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&destination, "destination", "d", "rule_files", "Path to the directory to save the files in (default current working directory). If the directory doesn't exist, then it will be automatically created.")
	cmd.Flags().BoolVarP(&cloud, "cloud", "c", false, "Only download cloud rules")
	cmd.Flags().BoolVarP(&connector, "connector", "n", false, "Only download connector rules")

	return cmd
}

func saveCloudXMLRules(apiClient *sailpoint.APIClient, description string, includeTypes []string, excludeTypes []string, options *map[string]beta.ObjectExportImportOptions, destination string) error {

	job, _, err := apiClient.Beta.SPConfigAPI.ExportSpConfig(context.TODO()).ExportPayload(beta.ExportPayload{Description: &description, IncludeTypes: includeTypes, ExcludeTypes: excludeTypes, ObjectOptions: options}).Execute()
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	for {
		response, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExportStatus(context.TODO(), job.JobId).Execute()
		if err != nil {
			return err
		}
		if response.Status == "NOT_STARTED" || response.Status == "IN_PROGRESS" {
			color.Yellow("Status: %s. checking again in 5 seconds", response.Status)
			time.Sleep(5 * time.Second)
		} else {
			switch response.Status {
			case "COMPLETE":
				exportData, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExport(context.TODO(), job.JobId).Execute()
				if err != nil {
					return err
				}

				for _, v := range exportData.Objects {
					if v.Object["type"] == nil || slices.Contains(cloudRuleTypes, v.Object["type"].(string)) {
						var RuleType string
						if v.Object["type"] == nil {
							RuleType = "Generic"
						} else {
							RuleType = v.Object["type"].(string)
						}

						// Create XML document
						doc := etree.NewDocument()
						doc.CreateProcInst("xml", `version='1.0' encoding='UTF-8'`)
						doc.CreateDirective("DOCTYPE Rule PUBLIC \"sailpoint.dtd\" \"sailpoint.dtd\"")

						// Create Rule element
						rule := doc.CreateElement("Rule")
						rule.CreateAttr("name", v.Object["name"].(string))
						rule.CreateAttr("type", RuleType)

						// Add Description if it exists
						if v.Object["description"] != nil {
							desc := rule.CreateElement("Description")
							desc.SetText(v.Object["description"].(string))
						}

						// Add Signature if it exists
						if len(v.Object["signature"].(map[string]interface{})["input"].([]interface{})) > 0 {
							signature := rule.CreateElement("Signature")
							signature.CreateAttr("returnType", "String")

							// Add Inputs
							inputs := signature.CreateElement("Inputs")
							for _, input := range v.Object["signature"].(map[string]interface{})["input"].([]interface{}) {
								arg := inputs.CreateElement("Argument")
								arg.CreateAttr("name", input.(map[string]interface{})["name"].(string))
								if input.(map[string]interface{})["type"] != nil {
									arg.CreateAttr("type", input.(map[string]interface{})["type"].(string))
								}
								if input.(map[string]interface{})["description"] != nil {
									desc := arg.CreateElement("Description")
									desc.SetText(input.(map[string]interface{})["description"].(string))
								}
							}

							// Add Returns if it exists
							if v.Object["signature"].(map[string]interface{})["output"] != nil {
								returns := signature.CreateElement("Returns")
								output := v.Object["signature"].(map[string]interface{})["output"]

								if outputs, ok := output.([]interface{}); ok {
									for _, output := range outputs {
										arg := returns.CreateElement("Argument")
										arg.CreateAttr("name", output.(map[string]interface{})["name"].(string))
										if output.(map[string]interface{})["type"] != nil {
											arg.CreateAttr("type", output.(map[string]interface{})["type"].(string))
										}
										if output.(map[string]interface{})["description"] != nil {
											desc := arg.CreateElement("Description")
											desc.SetText(output.(map[string]interface{})["description"].(string))
										}
									}
								} else if outputMap, ok := output.(map[string]interface{}); ok {
									arg := returns.CreateElement("Argument")
									arg.CreateAttr("name", outputMap["name"].(string))
									if outputMap["type"] != nil {
										arg.CreateAttr("type", outputMap["type"].(string))
									}
									if outputMap["description"] != nil {
										desc := arg.CreateElement("Description")
										desc.SetText(outputMap["description"].(string))
									}
								}
							}
						}

						// Add Source with CDATA
						source := rule.CreateElement("Source")
						source.CreateCharData("<![CDATA[\n" + v.Object["sourceCode"].(map[string]interface{})["script"].(string) + "\n]]>")

						// Write to file
						doc.Indent(2)
						xmlStr, err := doc.WriteToString()
						if err != nil {
							return err
						}
						// Replace encoded characters back to their original form
						xmlStr = strings.ReplaceAll(xmlStr, "&quot;", "\"")
						xmlStr = strings.ReplaceAll(xmlStr, "&amp;", "&")
						xmlStr = strings.ReplaceAll(xmlStr, "&lt;", "<")
						xmlStr = strings.ReplaceAll(xmlStr, "&gt;", ">")
						err = output.WriteFile(destination+"/cloud", "Rule - "+RuleType+" - "+v.Object["name"].(string)+".xml", []byte(xmlStr))
						if err != nil {
							return err
						}

						log.Info("Job Complete")
					}
				}

				return nil
			case "CANCELLED":
				return fmt.Errorf("export task cancelled")
			case "FAILED":
				return fmt.Errorf("export task failed")
			}
		}
	}
}

func saveJSONConnectorRules(apiClient *sailpoint.APIClient, destination string) error {

	connectorRules, _, err := apiClient.Beta.ConnectorRuleManagementAPI.GetConnectorRuleList(context.TODO()).Execute()

	if err != nil {
		return err
	}

	for _, v := range connectorRules {
		err := output.SaveJSONFile(v, v.Name+".json", destination+"/connector")
		if err != nil {
			return err
		}
	}

	log.Info("Job Complete")

	return nil
}
