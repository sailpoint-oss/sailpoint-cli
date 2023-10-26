// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sanitize

import (
	_ "embed"
	"os"
	"regexp"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed sanitize.md
var sanitizeHelp string

func NewSanitizeCommand() *cobra.Command {
	help := util.ParseHelp(sanitizeHelp)
	cmd := &cobra.Command{
		Use:     "sanitize",
		Short:   "Sanitize a file of sensitive data",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			for _, filePath := range args {
				// Open the file, read it in as plain text, and sanitize it of all access tokens
				file, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}

				tokens := regexp.MustCompile(`(ey[A-Za-z0-9-_=]+).(ey[A-Za-z0-9-_=]+).[A-Za-z0-9-_.+=]+`)
				origins := regexp.MustCompile(`\{[\s]+"name": "origin",[\s]+"value":[\s]+"[A-Za-z0-9:\/\-.]+"[\s]+\},`)

				redactedFile := origins.ReplaceAllString(tokens.ReplaceAllString(string(file), "REDACTED"), `{"name":"origin", "value":"REDACTED"},`)

				// Write the sanitized file back to disk
				err = os.WriteFile(filePath, []byte(redactedFile), 0644)
				if err != nil {
					return err
				}

			}

			return nil
		},
	}

	return cmd

}
