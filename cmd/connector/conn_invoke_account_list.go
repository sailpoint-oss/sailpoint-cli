// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnInvokeAccountListCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-list",
		Short: "Invoke a std:account:list command",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			var stateful *bool
			if s := cmd.Flags().Lookup("stateful"); s != nil {
				if s.Value.String() == "true" {
					t := true
					stateful = &t
				}
			}

			var stateId *string
			if si := cmd.Flags().Lookup("stateId"); si != nil {
				if siv := si.Value.String(); siv != "" {
					stateId = &siv
				}
			}

			var schema map[string]interface{}
			if sc := cmd.Flags().Lookup("schema"); sc != nil {
				if scv := sc.Value.String(); scv != "" {
					fmt.Println(scv)
					b, err := os.ReadFile(scv)
					if err != nil {
						return err
					}

					//schema, err = make(map[string]string)
					err = json.Unmarshal(b, &schema)
					if err != nil {
						return err
					}
				}
			}

			_, state, printable, err := cc.AccountList(ctx, stateful, stateId, schema)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Accounts:")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(printable))

			if state != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nState:\n")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(state))
			}
			return nil
		},
	}

	cmd.Flags().Bool("stateful", false, "Optional - Run command with state")
	cmd.Flags().String("stateId", "", "Optional - The state ID from a previous command invocation result")
	cmd.Flags().String("schema", "", "Optional - Custom account schema")

	return cmd
}
