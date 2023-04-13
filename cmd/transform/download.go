// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/transform"
	"github.com/spf13/cobra"
)

func newDownloadCmd() *cobra.Command {
	var destination string
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Download all Transforms from IdentityNow",
		Long:    "\nDownload all Transforms from IdentityNow\n\n",
		Example: "sail transform downlooad -d transform_files | sail transform dl",
		Aliases: []string{"dl"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			transforms, err := transform.GetTransforms()
			if err != nil {
				return err
			}

			err = transform.ListTransforms()
			if err != nil {
				return err
			}

			for _, v := range transforms {
				filename := strings.ReplaceAll(v.Name, " ", "") + ".json"
				content, _ := json.MarshalIndent(v, "", "    ")

				var err error

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

			config.Log.Info("Transforms downloaded successfully", "path", destination)

			return nil
		},
	}

	cmd.Flags().StringVarP(&destination, "destination", "d", "transform_files", "The path to the directory to save the files in (default current working directory).  If the directory doesn't exist, then it will be automatically created.")

	return cmd
}
