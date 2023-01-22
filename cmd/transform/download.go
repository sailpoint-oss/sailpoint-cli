// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newDownloadCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "download transforms",
		Long:    "Download transforms to local storage",
		Example: "sail transform dl -d transform_files|\nsail transform dl",
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

			destination := cmd.Flags().Lookup("destination").Value.String()

			err = listTransforms(client, endpoint, cmd)
			if err != nil {
				return err
			}

			for _, v := range transforms {
				filename := strings.ReplaceAll(v["name"].(string), " ", "") + ".json"
				content, _ := json.MarshalIndent(v, "", "    ")

				var err error
				if destination == "" {
					destination = "transform_files"
				}

				// Make sure the output dir exists first
				err = os.MkdirAll(destination, os.ModePerm)
				if err != nil {
					return err
				}

				// Make sure to create the files if they dont exist
				file, err := os.OpenFile((filepath.Join(destination, filename)), os.O_RDWR|os.O_CREATE, 0777)
				if err != nil {
					return err
				}
				_, err = file.Write(content)
				if err != nil {
					return err
				}

				if err != nil {
					return err
				}
			}

			color.Green("Transforms downloaded successfully to %v", destination)

			return nil
		},
	}

	cmd.Flags().StringP("destination", "d", "", "The path to the directory to save the files in (default current working directory).  If the directory doesn't exist, then it will be automatically created.")

	return cmd
}
