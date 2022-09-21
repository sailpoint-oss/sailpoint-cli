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

func newCreateCmd(client client.Client) *cobra.Command {

	// TODO: Clean up and not send display name
	type create struct {
		DisplayName string `json:"displayName"`
		Alias       string `json:"alias"`
	}

	cmd := &cobra.Command{
		Use:     "create <transform data>",
		Short:   "Create transform",
		Long:    "Create a transform",
		Example: "sp transforms create < transform.json",
		Aliases: []string{"c"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			var data map[string]interface{}

			err := json.NewDecoder(os.Stdin).Decode(&data)

			if err != nil {
				log.Fatal(err)
			}

			if data["name"] == nil {
				log.Fatal("The transform must have a name.")
				return nil
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

	return cmd
}
