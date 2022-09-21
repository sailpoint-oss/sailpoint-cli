// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/sailpoint-oss/sp-cli/client"
	"github.com/sailpoint-oss/sp-cli/util"
	"github.com/spf13/cobra"
)

func newUpdateCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <transform-data>",
		Short:   "Update transform",
		Long:    "Update a transform specified by the id in the input.",
		Example: "sp transforms update < /path/to/transform.json",
		Aliases: []string{"u"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var data map[string]interface{}

			err := json.NewDecoder(os.Stdin).Decode(&data)
			if err != nil {
				log.Fatal(err)
			}

			if data["id"] == nil {
				log.Fatal("The input must contain an id.")
				return nil
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

	return cmd
}
