// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package rule

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"slices"
	"time"

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

						//Make Rule XML Object
						rule := &Rule{}
						rule.Name = v.Object["name"].(string)
						rule.Type = RuleType

						if v.Object["description"] != nil {
							rule.Description = v.Object["description"].(string)
						} else {
							rule.Description = ""
						}

						rule.Source = "<![CDATA[\n" + v.Object["sourceCode"].(map[string]interface{})["script"].(string) + "\n]]>"

						var ruleSignature = &Signature{}

						//Add Signature if it exists
						if len(v.Object["signature"].(map[string]interface{})["input"].([]interface{})) > 0 {

							ruleSignature.Inputs = &Inputs{Argument: []Argument{}}
							for _, v := range v.Object["signature"].(map[string]interface{})["input"].([]interface{}) {
								argument := Argument{}

								argument.Name = v.(map[string]interface{})["name"].(string)

								if v.(map[string]interface{})["type"] != nil {
									argument.Type = v.(map[string]interface{})["type"].(string)
								} else {
									argument.Type = ""
								}

								if v.(map[string]interface{})["description"] != nil {
									argument.Description = v.(map[string]interface{})["description"].(string)
								} else {
									argument.Description = ""
								}

								ruleSignature.Inputs.Argument = append(ruleSignature.Inputs.Argument, argument)
							}

							rule.Signature = ruleSignature
						}

						if v.Object["signature"].(map[string]interface{})["output"] != nil {
							ruleSignature.Returns = &Returns{Argument: []Argument{}}

							if _, ok := v.Object["signature"].(map[string]interface{})["output"].([]interface{}); ok {
								for _, v := range v.Object["signature"].(map[string]interface{})["output"].([]interface{}) {

									argument := Argument{}

									argument.Name = v.(map[string]interface{})["name"].(string)

									if v.(map[string]interface{})["type"] != nil {
										argument.Type = v.(map[string]interface{})["type"].(string)
									} else {
										argument.Type = ""
									}

									if v.(map[string]interface{})["description"] != nil {
										argument.Description = v.(map[string]interface{})["description"].(string)
									} else {
										argument.Description = ""
									}

									ruleSignature.Returns.Argument = append(ruleSignature.Returns.Argument, argument)
								}

							} else {
								output := v.Object["signature"].(map[string]interface{})["output"].(map[string]interface{})

								argument := Argument{}

								argument.Name = output["name"].(string)

								if output["type"] != nil {
									argument.Type = output["type"].(string)
								} else {
									argument.Type = ""
								}

								if output["description"] != nil {
									argument.Description = output["description"].(string)
								} else {
									argument.Description = ""
								}

								ruleSignature.Returns.Argument = append(ruleSignature.Returns.Argument, argument)
							}
						}

						out, _ := xml.MarshalIndent(rule, "", "  ")

						out = []byte(xml.Header + SailPointHeader + string(out))

						out = bytes.Replace(out, []byte("&#xA;"), []byte("\n"), -1)
						out = bytes.Replace(out, []byte("&#xD;"), []byte("\r"), -1)
						out = bytes.Replace(out, []byte("&#34;"), []byte("\""), -1)
						out = bytes.Replace(out, []byte("&amp;"), []byte("&"), -1)
						out = bytes.Replace(out, []byte("&#x9;"), []byte("\t"), -1)
						out = bytes.Replace(out, []byte("&lt;"), []byte("<"), -1)
						out = bytes.Replace(out, []byte("&gt;"), []byte(">"), -1)

						err := output.WriteFile(destination+"/cloud", "Rule - "+rule.Type+" - "+rule.Name+".xml", out)

						if err != nil {
							return err
						}

						log.Info("Job Complete")
					}
				}

				if err != nil {
					return err
				}
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
