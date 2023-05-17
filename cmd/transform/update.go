// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/beta"
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
			var transform sailpointbetasdk.Transform

			filepath := cmd.Flags().Lookup("file").Value.String()
			if filepath != "" {
				file, err := os.Open(filepath)
				if err != nil {
					return err
				}
				defer file.Close()

				err = json.NewDecoder(file).Decode(&transform)
				if err != nil {
					return err
				}
			} else {
				err := json.NewDecoder(os.Stdin).Decode(&transform)
				if err != nil {
					return err
				}
			}

			if transform.Id == nil {
				return fmt.Errorf("the input must contain an id")
			}

			id := transform.Id
			transform.Id = nil // ID can't be present in the update payload

			log.Info("Updating Transform", "transformID", id)

			updateTransform := sailpointbetasdk.NewTransform(transform.Name, transform.Type, transform.Attributes)

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			_, resp, err := apiClient.Beta.TransformsApi.UpdateTransform(context.TODO(), *id).Transform(*updateTransform).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
