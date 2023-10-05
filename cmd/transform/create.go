// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	var filepath string
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create an IdentityNow Transform from a file",
		Long:    "\nCreate an IdentityNow Transform from a file\n\n",
		Example: "sail transform c -f /path/to/transform.json\nsail transform c < /path/to/transform.json\necho /path/to/transform.json | sail transform c",
		Aliases: []string{"c"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var transform beta.Transform
			var decoder *json.Decoder

			if filepath != "" {
				file, err := os.Open(filepath)
				if err != nil {
					return err
				}
				defer file.Close()
				decoder = json.NewDecoder(bufio.NewReader(file))
			} else {
				decoder = json.NewDecoder(bufio.NewReader(os.Stdin))
			}

			if err := decoder.Decode(&transform); err != nil {
				return err
			}

			log.Debug("Filepath", "path", filepath)

			log.Debug("Transform", "transform", transform)

			if transform.GetName() == "" {
				return fmt.Errorf("the transform must have a name")
			}

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			transformObj, resp, err := apiClient.Beta.TransformsApi.CreateTransform(ctx).Transform(transform).Execute()
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			log.Info("Transform created successfully")

			cmd.Print(transformObj.Id)

			return nil
		},
	}

	cmd.Flags().StringVarP(&filepath, "file", "f", "", "The path to the transform file")

	return cmd
}
