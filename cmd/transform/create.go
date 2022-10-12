// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/sailpoint-oss/sailpoint-cli/util"
	"github.com/spf13/cobra"
)

func newCreateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create transform",
		Long:    "Create a transform from a file [-f] or standard input (if no file is specified).",
		Example: "sp transforms create -f /path/to/transform.json",
		Aliases: []string{"c"},
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

			if data["name"] == nil {
				return fmt.Errorf("The transform must have a name.")
			}

			raw, err := json.Marshal(data)
			if err != nil {
				return err
			}

			endpoint := cmd.Flags().Lookup("transforms-endpoint").Value.String()
			resp, err := client.Post(cmd.Context(), util.ResourceUrl(endpoint), "application/json", bytes.NewReader(raw))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("create transform failed. status: %s\nbody: %s", resp.Status, body)
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
