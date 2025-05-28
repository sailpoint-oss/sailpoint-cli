// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newCustomizerCreateVersionCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upload",
		Short:   "Upload connector customizer",
		Example: "sail conn customizers upload -c 1234 -f path/to/zip/archive.zip",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := cmd.Flags().Lookup("id").Value.String()
			archivePath := cmd.Flags().Lookup("file").Value.String()

			f, err := os.Open(archivePath)
			if err != nil {
				return err
			}
			defer f.Close()

			info, err := f.Stat()
			if err != nil {
				return err
			}

			_, err = zip.NewReader(f, info.Size())
			if err != nil {
				return err
			}

			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}

			resp, err := client.Post(cmd.Context(), util.ResourceUrl(connectorCustomizersEndpoint, id, "versions"), "application/zip", f, nil)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("upload customizer failed. status: %s\nbody: %s", resp.Status, string(body))
			}

			var cv customizerVersion
			err = json.NewDecoder(resp.Body).Decode(&cv)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader(customizerVersionColumns)
			table.Append(cv.columns())
			table.Render()

			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Connector customizer ID")
	_ = cmd.MarkFlagRequired("id")

	cmd.Flags().StringP("file", "f", "", "ZIP Archive")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}
