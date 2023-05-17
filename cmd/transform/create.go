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
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create an IdentityNow Transform from a file",
		Long:    "\nCreate an IdentityNow Transform from a file\n\n",
		Example: "sail transform c -f /path/to/transform.json\nsail transform c < /path/to/transform.json\necho /path/to/transform.json | sail transform c",
		Aliases: []string{"c"},
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

			log.Debug("Transform", "transform", util.PrettyPrint(transform))

			if transform.Name == "" {
				return fmt.Errorf("the transform must have a name")
			}

			if transform.Id != nil {
				return fmt.Errorf("the transform cannot have an ID")
			}

			createTransform := sailpointbetasdk.NewTransform(transform.Name, transform.Type, transform.Attributes)

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			transformObj, resp, err := apiClient.Beta.TransformsApi.CreateTransform(context.TODO()).Transform(*createTransform).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			log.Info("Transform created successfully")

			cmd.Print(*transformObj.Id)

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
