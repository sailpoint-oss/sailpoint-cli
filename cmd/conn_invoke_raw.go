package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

type rawCommand struct {
	Type  string          `json:"type"`
	Input json.RawMessage `json:"input"`
}

func newConnInvokeRaw(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   `raw < command.json`,
		Short: "Invoke a raw command",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			raw := &rawCommand{}
			filePath := cmd.Flags().Lookup("file").Value.String()

			var reader io.Reader
			if len(filePath) > 0 {
				reader, err = os.Open(filePath)
				if err != nil {
					return err
				}
			} else {
				reader = os.Stdin
			}
			decoder := json.NewDecoder(reader)
			err = decoder.Decode(raw)
			if err != nil {
				return err
			}

			rawResponse, err := cc.Invoke(ctx, raw.Type, raw.Input)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))

			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "JSON file containing a command")

	return cmd
}
