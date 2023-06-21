// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package initialize

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const packageJsonName = "package.json"

func InitializeProject(templateContents embed.FS, templateDirName string, projName string) error {

	if projName == "" {
		return errors.New("connector name cannot be empty")

	}

	if f, err := os.Stat(projName); err == nil && f.IsDir() && f.Name() == projName {
		return fmt.Errorf("error: project '%s' already exists", projName)

	}

	if err := createDir(projName); err != nil {
		_ = os.RemoveAll(projName)
		return err
	}

	err := fs.WalkDir(templateContents, templateDirName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == templateDirName {
			return nil
		}

		if d.IsDir() {
			if err := createDir(filepath.Join(projName, d.Name())); err != nil {
				return err
			}
		} else {
			fileName := filepath.Join(projName, strings.TrimPrefix(path, templateDirName))

			data, err := templateContents.ReadFile(path)
			if err != nil {
				return err
			}

			if err := createFile(fileName, data); err != nil {
				return err
			}
		}

		if d.Name() == packageJsonName {
			fileAbsPath, err := filepath.Abs(filepath.Join(projName, strings.TrimPrefix(path, templateDirName)))
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
		return err
	}

	printDir(projName, 0)
	return nil
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

// printFile prints file branch
func printFile(name string, depth int) {
	_, _ = fmt.Printf("%s|-- %s\n", strings.Repeat("|   ", depth), filepath.Base(name))
}

// printDir prints directory tree from specified path
func printDir(path string, depth int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		_, _ = fmt.Printf("error reading %s: %v", path, err)
		return
	}

	printFile(path, depth)
	for _, entry := range entries {
		if entry.IsDir() {
			printDir(filepath.Join(path, entry.Name()), depth+1)
		} else {
			printFile(entry.Name(), depth+1)
		}
	}
}
