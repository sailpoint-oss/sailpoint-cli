// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/sailpoint-oss/sp-cli/client"
	"github.com/spf13/cobra"
)

func newTransformDownloadCmd(client client.Client) *cobra.Command {
	return &cobra.Command{
		Use:     "download",
		Short:   "Download Transforms",
		Long:    "Download Transforms To Local Storage",
		Aliases: []string{"dl"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("transforms-endpoint").Value.String()

			resp, err := client.Get(cmd.Context(), endpoint)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("non-200 response: %s\nbody: %s", resp.Status, body)
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			// Since we just want to save the content to files, we don't need
			// to parse individual fields.  Just get the string representation.
			var transforms []map[string]interface{}

			err = json.Unmarshal(raw, &transforms)
			if err != nil {
				return err
			}

			for _, v := range transforms {
				filename := v["name"].(string) + ".json"
				content, _ := json.MarshalIndent(v, "", "    ")
				err := ioutil.WriteFile(filename, content, os.ModePerm)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}
