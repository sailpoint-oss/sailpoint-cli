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
	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	var filepath string
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a transform from a file",
		Long:    "\nCreate a new transform in Identity Security Cloud from a JSON file.\nThe file can be specified with the --file flag, piped via stdin,\nor redirected from a file.\n",
		Example: "  sail transform create -f /path/to/transform.json\n  sail transform create < /path/to/transform.json",
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

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			transformObj, resp, err := apiClient.Beta.TransformsAPI.CreateTransform(ctx).Transform(transform).Execute()
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
