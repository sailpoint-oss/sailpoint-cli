package connector

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	connvalidate "github.com/sailpoint-oss/sailpoint-cli/cmd/connector/validate"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
)

const (
	accountReadLimit = 8
)

func newConnValidateCmd(apiClient client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate connector behavior",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Check if we just need to list checks
			list, _ := strconv.ParseBool(cmd.Flags().Lookup("list").Value.String())
			if list {
				table := tablewriter.NewWriter(os.Stdout)
				table.Header([]any{"ID", "Description"}...)
				for _, c := range connvalidate.Checks {
					table.Append([]string{
						c.ID,
						c.Description,
					})
				}
				table.Render()
				return nil
			}

			cc, err := connClient(cmd, apiClient)
			if err != nil {
				return err
			}

			check := cmd.Flags().Lookup("check").Value.String()

			isReadOnly, _ := strconv.ParseBool(cmd.Flags().Lookup("read-only").Value.String())
			readLimitVal, err := getReadLimitVal(cmd)
			if err != nil {
				return fmt.Errorf("invalid value of readLimit: %v", err)
			}

			valid := connvalidate.NewValidator(connvalidate.Config{
				Check:     check,
				ReadOnly:  isReadOnly,
				ReadLimit: readLimitVal,
			}, cc)

			results, err := valid.Run(ctx)
			if err != nil {
				return err
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.Header([]any{"ID", "Result", "Errors", "Warnings", "Skipped"}...)
			hasFailedCheck := false
			for _, res := range results {
				var result = aurora.Green("PASS")
				if len(res.Errors) > 0 {
					hasFailedCheck = true
					result = aurora.Red("FAIL")
				}

				if len(res.Skipped) > 0 {
					result = aurora.Yellow("SKIPPED")
				}

				table.Append([]string{
					aurora.Blue(res.ID).String(),
					result.String(),
					aurora.Red(strings.Join(res.Errors, "\n\n")).String(),
					aurora.Yellow(strings.Join(res.Warnings, "\n\n")).String(),
					aurora.Yellow(strings.Join(res.Skipped, "\n\n")).String(),
				})
			}
			table.Render()

			if hasFailedCheck {
				return fmt.Errorf("at least one check has failed")
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringP("check", "", "", "Run a specific check")
	cmd.PersistentFlags().BoolP("list", "l", false, "List checks; don't run checks")
	cmd.PersistentFlags().BoolP("read-only", "r", false, "Run all checks that don't modify connector's data")

	cmd.PersistentFlags().StringP("version", "v", "", "Run against a specific version")
	cmd.MarkFlagRequired("version")

	cmd.PersistentFlags().StringP("config-path", "p", "", "Path to config to use for test command")
	cmd.MarkFlagRequired("config-path")

	cmd.PersistentFlags().StringP("id", "c", "", "Connector ID or Alias")
	cmd.MarkFlagRequired("id")

	return cmd
}

func getReadLimitVal(cmd *cobra.Command) (int64, error) {
	readLimitVal, err := cmd.Flags().GetInt64("read-limit")
	if err != nil {
		return 0, err
	}
	if readLimitVal <= 0 {
		return 0, fmt.Errorf("readLimit value cannot be smaller than or equal to 0")
	}
	return readLimitVal, nil
}
