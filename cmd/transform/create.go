// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/sdk-output/v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "create transform",
		Long:    "Create a transform from a file [-f] or standard input (if no file is specified).",
		Example: "sail transform c -f /path/to/transform.json\nsail transform c < /path/to/transform.json\necho /path/to/transform.json | sail transform c",
		Aliases: []string{"c"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var data map[string]interface{}

			err := config.InitConfig()
			if err != nil {
				return err
			}

			filepath := cmd.Flags().Lookup("file").Value.String()
			if filepath != "" {
				file, err := os.Open(filepath)
				if err != nil {
					return err
				}
				defer file.Close()

				err = json.NewDecoder(file).Decode(&data)
				if err != nil {
					return err
				}
			} else {
				err := json.NewDecoder(os.Stdin).Decode(&data)
				if err != nil {
					return err
				}
			}

			if data["name"] == nil {
				return fmt.Errorf("the transform must have a name")
			}

			if data["id"] != nil {
				return fmt.Errorf("the transform cannot have an ID")
			}

			transform := sailpointsdk.NewTransform(data["name"].(string), data["type"].(string), data["attributes"].(map[string]interface{}))

			apiClient := config.InitAPIClient()

			transformObj, resp, err := apiClient.V3.TransformsApi.CreateTransform(context.TODO()).Transform(*transform).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			color.Green("Transform created successfully")

			cmd.Print(*transformObj.Id)

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
