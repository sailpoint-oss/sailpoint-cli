// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package spconfig

import (
	"context"
	"encoding/json"
	"os"

	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/sdk-output/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/spconfig"
	"github.com/spf13/cobra"
)

func newImportCommand() *cobra.Command {
	var filepath string
	var payload *sailpointbetasdk.ImportOptions
	cmd := &cobra.Command{
		Use:     "import",
		Short:   "begin an import job in identitynow",
		Long:    "initiate an import job in identitynow",
		Example: "sail spconfig import",
		Aliases: []string{"imp"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			file, err := os.Open(filepath)
			if err != nil {
				return err
			}
			defer file.Close()

			err = json.NewDecoder(file).Decode(&payload)
			if err != nil {
				return err
			}

			ctx := context.TODO()

			payload = sailpointbetasdk.NewImportOptions()

			job, _, err := apiClient.Beta.SPConfigApi.ImportSpConfig(ctx).Data(args[0]).Options(*payload).Execute()
			if err != nil {
				return err
			}

			spconfig.PrintJob(*job)

			return nil
		},
	}

	cmd.Flags().StringVarP(&filepath, "filePath", "f", "", "the path to the file containing the import payload")
	cmd.MarkFlagRequired("filepath")

	return cmd
}
