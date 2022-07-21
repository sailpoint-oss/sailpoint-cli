// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statColumns = []string{"Command", "Invocation Count", "Error Count", "Error Rate", "Elapsed Avg", "Elapsed 95th Percentile"}

const (
	day  = int64(24 * time.Hour)
	week = int64(7 * 24 * time.Hour)
)

var durationMap = map[byte]int64{
	'd': day,
	'w': week,
}

func newConnStatsCmd(spClient client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats",
		Short:   "Command Stats",
		Long:    "Command execution stats for a tenant, default to last 24hs if duration not specified",
		Example: "sp conn stats",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := getTenantStats(spClient, cmd); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringP("stats-endpoint", "o", viper.GetString("baseurl")+client.StatsEndpoint, "Override stats endpoint")
	cmd.Flags().StringP("duration", "d", "", `Length of time represented by an integer(1-9) and a duration unit. Supported duration units: d,w. eg 1d, 3w`)
	cmd.Flags().StringP("id", "c", "", "Connector ID")
	return cmd
}

func getTenantStats(spClient client.Client, cmd *cobra.Command) error {
	endpoint := cmd.Flags().Lookup("stats-endpoint").Value.String()
	lc := client.NewLogsClient(spClient, endpoint)

	connectorID := cmd.Flags().Lookup("id").Value.String()
	durationStr := cmd.Flags().Lookup("duration").Value.String()
	duration, err := parseDuration(durationStr)
	if err != nil {
		return err
	}
	if duration == nil {
		return fmt.Errorf("invalid duration")
	}

	from := time.Now().Add(-*duration)
	tenantStats, err := lc.GetStats(cmd.Context(), from, connectorID)
	if err != nil {
		return err
	}
	for _, c := range tenantStats.ConnectorStats {
		table := tablewriter.NewWriter(cmd.OutOrStdout())
		table.SetHeader(statColumns)
		connAlias := ""
		if c.ConnectorAlias != "" {
			connAlias = fmt.Sprintf("(%v)", c.ConnectorAlias)
		}
		connTitle := fmt.Sprintf("Connector : %v %s ", c.ConnectorID, connAlias)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), connTitle)
		for _, v := range c.Stats {
			table.Append(v.Columns())
		}
		table.Render()
		fmt.Fprintln(cmd.OutOrStdout())
	}

	return nil
}

func parseDuration(durationStr string) (*time.Duration, error) {
	defaultDuration := time.Duration(day)
	if len(durationStr) == 0 {
		return &defaultDuration, nil
	}
	if !validDuration(durationStr) {
		return nil, fmt.Errorf("invalid duration")
	}
	durationNum := int64(durationStr[0] - '0')
	duration := time.Duration(durationNum * durationMap[durationStr[1]])
	return &duration, nil
}

func validDuration(durationStr string) bool {
	if len(durationStr) != 2 {
		return false
	}
	// The first character must be [1-9]
	if !('1' <= durationStr[0] && durationStr[0] <= '9') {
		return false
	}
	// The second character must be on of the supported duration[d,w]
	if _, ok := durationMap[durationStr[1]]; !ok {
		return false
	}
	return true
}
