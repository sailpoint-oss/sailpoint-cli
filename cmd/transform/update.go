// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update a transform from a file",
		Long:    "\nUpdate an existing transform in Identity Security Cloud from a JSON file.\nThe file can be specified with the --file flag, piped via stdin,\nor redirected from a file.\n",
		Example: "  sail transform update -f /path/to/transform.json\n  sail transform update < /path/to/transform.json",
		Aliases: []string{"u"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var transform beta.TransformRead

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

			if transform.Id == "" {
				return fmt.Errorf("the input must contain an id")
			}

			id := transform.Id
			// ID, Internal, Name, and Type can't be present in the update payload

			log.Info("Updating Transform", "transformID", id)

			updateTransform := beta.Transform{Attributes: transform.Attributes, Type: transform.Type, Name: transform.Name}

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			_, resp, err := apiClient.Beta.TransformsAPI.UpdateTransform(context.TODO(), id).Transform(updateTransform).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Path to the transform file")

	return cmd
}
