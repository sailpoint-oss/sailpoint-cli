// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnCreateVersionCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload Connector",
		Long:  "Upload Connector",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			connectorRef := cmd.Flags().Lookup("id").Value.String()
			archivePath := cmd.Flags().Lookup("file").Value.String()
			tagName := cmd.Flags().Lookup("tag").Value.String()

			f, err := os.Open(archivePath)
			if err != nil {
				return err
			}

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

			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()
			resp, err := client.Post(cmd.Context(), connResourceUrl(endpoint, connectorRef, "versions"), "application/zip", f)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("upload failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			var v connectorVersion
			err = json.Unmarshal(raw, &v)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader(connectorVersionColumns)
			table.Append(v.columns())
			table.Render()

			if tagName != "" {
				resp, err := client.Get(cmd.Context(), connResourceUrl(endpoint, connectorRef, "tags", tagName))
				if err != nil {
					return err
				}
				defer func() {
					_ = resp.Body.Close()
				}()

				// If tag exists, update the tag with new version.
				// Otherwise create the tag
				if resp.StatusCode == http.StatusOK {
					err = updateTagWithVersion(cmd, client, endpoint, connectorRef, tagName, uint32(v.Version))
				} else {
					err = createTagWithVersion(cmd, client, endpoint, connectorRef, tagName, uint32(v.Version))
				}

				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("id", "c", "", "Connector ID or Alias")
	_ = cmd.MarkFlagRequired("id")

	cmd.Flags().StringP("file", "f", "", "ZIP Archive")
	_ = cmd.MarkFlagRequired("file")

	cmd.Flags().StringP("tag", "t", "", "Update a tag with this version. Tag will be created if not exist. (Optional)")

	bindDevConfig(cmd.Flags())

	return cmd
}

// updateTagWithVersion updates an exiting tag with a new version of connector code
func updateTagWithVersion(cmd *cobra.Command, client client.Client, endpoint string, connectorID string, tagName string, version uint32) error {
	raw, err := json.Marshal(TagUpdate{ActiveVersion: version})
	if err != nil {
		return err
	}

	resp, err := client.Put(cmd.Context(), connResourceUrl(endpoint, connectorID, "tags", tagName), "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update connector tag failed. status: %s\nbody: %s", resp.Status, body)
	}

	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var t tag
	err = json.Unmarshal(raw, &t)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetHeader(tagColumns)
	table.Append(t.columns())
	table.Render()
	return nil
}

// createTagWithVersion creates a tag pointing to a version of connector code
func createTagWithVersion(cmd *cobra.Command, client client.Client, endpoint string, connectorID string, tagName string, version uint32) error {
	raw, err := json.Marshal(TagCreate{TagName: tagName, ActiveVersion: version})
	if err != nil {
		return err
	}

	resp, err := client.Post(cmd.Context(), connResourceUrl(endpoint, connectorID, "tags"), "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create connector tag failed. status: %s\nbody: %s", resp.Status, body)
	}

	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var t tag
	err = json.Unmarshal(raw, &t)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetHeader(tagColumns)
	table.Append(t.columns())
	table.Render()

	return nil
}
