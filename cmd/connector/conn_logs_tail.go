// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"time"

	connclient "github.com/sailpoint-oss/sailpoint-cli/cmd/connector/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnLogsTailCmd(client client.Client) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "tail",
		Short:   "Tail Logs",
		Example: "sail logs tail",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := tailLogs(client, cmd); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}

func tailLogs(spClient client.Client, cmd *cobra.Command) error {
	handleLogs := func(logEvents *connclient.LogEvents, cmd *cobra.Command) error {
		if err := printLogs(logEvents, cmd); err != nil {
			return err
		}
		for _, l := range logEvents.Logs {
			updateLastSeenTime(l.Timestamp)
		}
		return nil
	}

	for {
		logInput.Filter.StartTime = nextFromTime()
		if err := getAllLogs(spClient, cmd, handleLogs); err != nil {
			return err
		}
		time.Sleep(2 * time.Second)
	}
}

var lastSeenTime *int64

func updateLastSeenTime(ts time.Time) {
	nextTimeMilli := ts.UnixMilli()
	if lastSeenTime == nil || nextTimeMilli > *lastSeenTime {
		lastSeenTime = &nextTimeMilli
	}
}

func nextFromTime() *time.Time {
	from := time.Now().Add(-5 * time.Minute)
	if lastSeenTime != nil {
		//to fetch from next millisecond
		from = time.UnixMilli(*lastSeenTime + 1)
	}
	return &from
}
