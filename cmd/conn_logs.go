// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logInput = client.LogInput{}

func newConnLogsCmd(spClient client.Client) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "logs",
		Short:   "List Logs",
		Example: "sp logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := formatDates(cmd); err != nil {
				return err
			}
			if logInput.Filter.StartTime == nil {
				from := time.Now().Add(-1 * time.Hour)
				logInput.Filter.StartTime = &from
			}
			if err := getAllLogs(spClient, cmd, printLogs); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringP("logs-endpoint", "o", viper.GetString("baseurl")+client.LogsEndpoint, "Override logs endpoint")
	//date filters
	cmd.Flags().StringP("start", "s", "", `start time - get the logs from this point. An absolute timestamp in RFC3339 format, or a relative time (eg. 2h). Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	cmd.Flags().StringP("stop", "", "", `end time - get the logs upto this point. An absolute timestamp in RFC3339 format, or a relative time (eg. 2h). Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`)
	//other filters
	cmd.PersistentFlags().StringVar(&logInput.Filter.Component, "component", "", "component type")
	cmd.PersistentFlags().StringVar(&logInput.Filter.TargetID, "target-id", "", "id of the specific target object")
	cmd.PersistentFlags().StringVar(&logInput.Filter.TargetName, "target-name", "", "name of the specifiy target")
	cmd.PersistentFlags().StringVar(&logInput.Filter.RequestID, "request-id", "", "associated request id")
	cmd.PersistentFlags().StringVar(&logInput.Filter.Event, "event", "", "event name")
	cmd.PersistentFlags().StringSliceVar(&logInput.Filter.LogLevels, "level", nil, "log levels")
	cmd.PersistentFlags().BoolP("raw", "r", false, "")

	cmd.AddCommand(newConnLogsTailCmd(spClient))

	return cmd
}

func printLogs(logEvents *client.LogEvents, cmd *cobra.Command) error {
	rawPrint, _ := cmd.Flags().GetBool("raw")

	if rawPrint {
		for _, t := range logEvents.Logs {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), t.RawString())
		}
	} else {
		for _, t := range logEvents.Logs {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), formatLog(t))
		}
	}
	return nil
}

func getAllLogs(spClient client.Client, cmd *cobra.Command, fn func(logEvents *client.LogEvents, cmd *cobra.Command) error) error {
	endpoint := cmd.Flags().Lookup("logs-endpoint").Value.String()
	lc := client.NewLogsClient(spClient, endpoint)

	logInput.NextToken = ""
	for {
		logEvents, err := lc.GetLogs(cmd.Context(), logInput)
		if err != nil {
			return err
		}
		if err := fn(logEvents, cmd); err != nil {
			return err
		}
		if logEvents.NextToken == nil {
			break
		} else {
			logInput.NextToken = *logEvents.NextToken
		}
	}
	return nil
}

func formatDates(cmd *cobra.Command) error {
	now := time.Now()
	startTimeFlag := cmd.Flags().Lookup("start").Value.String()
	stopTimeFlag := cmd.Flags().Lookup("stop").Value.String()

	if stopTimeFlag != "" && startTimeFlag == "" {
		return fmt.Errorf(`must provide a "--start" time when "--stop" specified`)
	}
	if startTimeFlag != "" {
		retTime, err := client.ParseTime(startTimeFlag, now)
		if err != nil {
			return err
		}
		logInput.Filter.StartTime = &retTime
	}
	if stopTimeFlag != "" {
		retTime, err := client.ParseTime(stopTimeFlag, now)
		if err != nil {
			return err
		}
		logInput.Filter.EndTime = &retTime
	}
	return nil
}

//Format log message for display
func formatLog(logMessage client.LogMessage) string {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	logLevelColor := color.New(color.FgHiWhite).SprintFunc()
	if logMessage.Level == "ERROR" {
		logLevelColor = color.New(color.FgHiRed).SprintFunc()
	}

	return fmt.Sprintf("%s%s%s%s", green(fmt.Sprintf("[%s]", logMessage.TimestampFormatted())),
		logLevelColor(fmt.Sprintf(" %-5s |", logMessage.Level)),
		yellow(fmt.Sprintf(" %-16s", logMessage.Event)),
		logLevelColor(fmt.Sprintf(" ▶︎ %s", logMessage.MessageString())))
}
