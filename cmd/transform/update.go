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

func newUpdateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update transform",
		Long:    "Update a transform from a file [-f] or standard input (if no file is specified).",
		Example: "sp transforms update -f /path/to/transform.json",
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
				return fmt.Errorf("The input must contain an id.")
			}

			id := data["id"].(string)
			delete(data, "id") // ID can't be present in the update payload

			raw, err := json.Marshal(data)
			if err != nil {
				return err
			}

			endpoint := cmd.Flags().Lookup("transforms-endpoint").Value.String()
			resp, err := client.Put(cmd.Context(), util.ResourceUrl(endpoint, id), "application/json", bytes.NewReader(raw))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("update transform failed. status: %s\nbody: %s", resp.Status, body)
			}

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "The path to the transform file")

	return cmd
}
