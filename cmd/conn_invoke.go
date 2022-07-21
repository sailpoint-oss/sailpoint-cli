// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

const (
	stdAccountCreate   = "std:account:create"
	stdAccountList     = "std:account:list"
	stdAccountRead     = "std:account:read"
	stdAccountUpdate   = "std:account:update"
	stdAccountDelete   = "std:account:delete"
	stdEntitlementList = "std:entitlement:list"
	stdEntitlementRead = "std:entitlement:read"
	stdTestConnection  = "std:test-connection"
)

func newConnInvokeCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoke",
		Short: "Invoke Command on a connector",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.PersistentFlags().StringP("version", "v", "", "Optional. Run against a specific version if provided. Otherwise run against the latest tag.")

	cmd.PersistentFlags().StringP("config-path", "p", "", "Path to config to use for commands")
	cmd.PersistentFlags().StringP("config-json", "", "", "Config JSON to use for commands")

	cmd.PersistentFlags().StringP("id", "c", "", "Connector ID or Alias")
	_ = cmd.MarkPersistentFlagRequired("id")

	cmd.AddCommand(
		newConnInvokeTestConnectionCmd(client),
		newConnInvokeAccountCreateCmd(client),
		newConnInvokeAccountDiscoverSchemaCmd(client),
		newConnInvokeAccountListCmd(client),
		newConnInvokeAccountReadCmd(client),
		newConnInvokeAccountUpdateCmd(client),
		newConnInvokeAccountDeleteCmd(client),
		newConnInvokeEntitlementListCmd(client),
		newConnInvokeEntitlementReadCmd(client),
		newConnInvokeRaw(client),
	)

	bindDevConfig(cmd.PersistentFlags())

	return cmd
}

func invokeConfig(cmd *cobra.Command) (json.RawMessage, error) {
	if cmd.Flags().Lookup("config-path").Value.String() == "" && cmd.Flags().Lookup("config-json").Value.String() == "" {
		return nil, fmt.Errorf("Either config-path or config-json must be set")
	}

	if cmd.Flags().Lookup("config-json") != nil && cmd.Flags().Lookup("config-json").Value.String() != "" {
		return json.RawMessage(cmd.Flags().Lookup("config-json").Value.String()), nil
	}

	return os.ReadFile(cmd.Flags().Lookup("config-path").Value.String())
}

type invokeCommand struct {
	ConnectorID string `json:"connectorID"`
	Version     *int   `json:"version"`

	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
	Input  json.RawMessage `json:"input"`
}

func connClient(cmd *cobra.Command, spClient client.Client) (*client.ConnClient, error) {
	connectorRef := cmd.Flags().Lookup("id").Value.String()
	version := cmd.Flags().Lookup("version").Value.String()
	endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()

	var v *int
	if version != "" {
		ver, err := strconv.Atoi(version)
		if err != nil {
			return nil, err
		}
		v = &ver
	}

	cfg, err := invokeConfig(cmd)
	if err != nil {
		return nil, err
	}
	cc := client.NewConnClient(spClient, v, cfg, connectorRef, endpoint)

	return cc, nil
}

func connClientWithCustomParams(spClient client.Client, cfg json.RawMessage, connectorID, version, endpoint string) (*client.ConnClient, error) {
	v, err := strconv.Atoi(version)
	if err != nil {
		return nil, err
	}

	cc := client.NewConnClient(spClient, &v, cfg, connectorID, endpoint)

	return cc, nil
}
