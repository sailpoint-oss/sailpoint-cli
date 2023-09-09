// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const customizerDirName = "static/customizer"

func newCustomizerInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init <customizer-name>",
		Short:   "Initialize new connector customizer project",
		Long:    `init sets up a new TypeScript project with sample connector customizer included for reference.`,
		Example: "sail conn customizers init \"My Connector\"",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projName := args[0]
			if projName == "" {
				printError(cmd.ErrOrStderr(), errors.New("connector customizer name cannot be empty"))
				return
			}

			if f, err := os.Stat(projName); err == nil && f.IsDir() && f.Name() == projName {
				printError(cmd.ErrOrStderr(), fmt.Errorf("Error: project '%s' already exists.\n", projName))
				return
			}

			if err := createDir(projName); err != nil {
				_ = os.RemoveAll(projName)
				printError(cmd.ErrOrStderr(), err)
				return
			}

			err := fs.WalkDir(staticDir, customizerDirName, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.Name() == customizerDirName {
					return nil
				}

				if d.IsDir() {
					if err := createDir(filepath.Join(projName, d.Name())); err != nil {
						return err
					}
				} else {
					fileName := filepath.Join(projName, strings.TrimPrefix(path, customizerDirName))

					data, err := staticDir.ReadFile(path)
					if err != nil {
						return err
					}

					if err := createFile(fileName, data); err != nil {
						return err
					}
				}

				if d.Name() == packageJsonName {
					fileAbsPath, err := filepath.Abs(filepath.Join(projName, strings.TrimPrefix(path, customizerDirName)))
					if err != nil {
						return err
					}

					if err := createFileFromTemplate(projName, d.Name(), fileAbsPath); err != nil {
						return err
					}
					return nil
				}

				return nil
			})
			if err != nil {
				_ = os.RemoveAll(projName)
				printError(cmd.ErrOrStderr(), err)
				return
			}

			printDir(cmd.OutOrStdout(), projName, 0)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully created project '%s'.\nRun `npm install` to install dependencies.\n", projName)
		},
	}

	return cmd
}
