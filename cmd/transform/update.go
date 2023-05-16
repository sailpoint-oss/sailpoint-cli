// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update a Transform in IdentityNow from a File",
		Long:    "\nUpdate a Transform in IdentityNow from a File\n\n",
		Example: "sail transform update --file ./assets/demo_update.json\nsail transform u -f /path/to/transform.json\nsail transform u < /path/to/transform.json\necho /path/to/transform.json | sail transform u",
		Aliases: []string{"u"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var data map[string]interface{}

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

			if data["id"] == nil {
				return fmt.Errorf("the input must contain an id")
			}

			id := data["id"].(string)
			delete(data, "id") // ID can't be present in the update payload

			log.Info("Updating Transaform", "transformID", id)

			transform := sailpointsdk.NewTransform(data["name"].(string), data["type"].(string), data["attributes"].(map[string]interface{}))

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			_, resp, err := apiClient.V3.TransformsApi.UpdateTransform(context.TODO(), id).Transform(*transform).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
