// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package rule

import (
	"context"
	"time"
	"encoding/json"
	"fmt"

	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
)

// var includeTypes = [...]string {"RULE"}

func newListCommand() *cobra.Command {

	var description = string ("Export of all rules")
	var objectOptions string
	var includeTypes = []string {"RULE"}
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

						// Save name and id to entries

						util.PrettyPrint(exportData)
						
						for _, v := range exportData.Objects {
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
