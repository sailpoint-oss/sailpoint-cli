// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package cmd

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed static/*
var staticDir embed.FS

const (
	staticDirName     = "static"
	packageJsonName   = "package.json"
	connectorSpecName = "connector-spec.json"
)

// newConnInitCmd is a connectors subcommand used to initialize a new connector project.
// It accepts one argument, project name, and generates appropriate directories and files
// to set up a working, sample project.
func newConnInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init <connector-name>",
		Short:   "Initialize new connector project",
		Long:    `init sets up a new TypeScript project with sample connector included for reference.`,
		Example: "sp connectors init \"My Connector\"",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projName := args[0]
			if projName == "" {
				printError(cmd.ErrOrStderr(), errors.New("connector name cannot be empty"))
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

			err := fs.WalkDir(staticDir, staticDirName, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.Name() == staticDirName {
					return nil
				}

				if d.IsDir() {
					if err := createDir(filepath.Join(projName, d.Name())); err != nil {
						return err
					}
				} else {
					fileName := filepath.Join(projName, strings.TrimPrefix(path, staticDirName))

					data, err := staticDir.ReadFile(path)
					if err != nil {
						return err
					}

					if err := createFile(fileName, data); err != nil {
						return err
					}
				}

				if d.Name() == packageJsonName || d.Name() == connectorSpecName {
					fileAbsPath, err := filepath.Abs(filepath.Join(projName, strings.TrimPrefix(path, staticDirName)))
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

// createFileFromTemplate fills the template file withe parameters,
// then creates the file in the target location
func createFileFromTemplate(projName string, filename string, fileAbsPath string) error {

	t, err := template.ParseFiles(fileAbsPath)
	if err != nil {
		return err
	}

	templateData := struct {
		ProjectName string
	}{
		ProjectName: projName,
	}

	packageJson := &bytes.Buffer{}
	if err := t.Execute(packageJson, templateData); err != nil {
		return err
	}

	if err := createFile(filepath.Join(projName, filename), packageJson.Bytes()); err != nil {
		return err
	}

	return nil
}

// createDir is a wrapper of os.MkdirAll, to generate project directories
func createDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	return nil
}

// createFile is a wrapper of os.WriteFile, to generate new source templates
func createFile(name string, data []byte) error {
	if err := os.WriteFile(name, data, 0644); err != nil {
		return err
	}

	return nil
}

// printError prints error in uniform format
func printError(w io.Writer, err error) {
	_, _ = fmt.Fprintf(w, "%v", err)
}

// printFile prints file branch
func printFile(w io.Writer, name string, depth int) {
	_, _ = fmt.Fprintf(w, "%s|-- %s\n", strings.Repeat("|   ", depth), filepath.Base(name))
}

// printDir prints directory tree from specified path
func printDir(w io.Writer, path string, depth int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		_, _ = fmt.Fprintf(w, "error reading %s: %v", path, err)
		return
	}

	printFile(w, path, depth)
	for _, entry := range entries {
		if entry.IsDir() {
			printDir(w, filepath.Join(path, entry.Name()), depth+1)
		} else {
			printFile(w, entry.Name(), depth+1)
		}
	}
}
