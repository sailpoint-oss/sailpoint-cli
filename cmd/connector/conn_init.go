// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package connector

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/go-git/go-git/v5"

	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
)

const githubAPI = "https://api.github.com/search/repositories"

type Repo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
}

type SearchResult struct {
	Items []Repo `json:"items"`
}

func fetchRepositories(org string, query string) ([]Repo, error) {
	searchQuery := fmt.Sprintf("org:%s %s", org, query)
	params := url.Values{}
	params.Add("q", searchQuery)
	apiURL := fmt.Sprintf("%s?%s", githubAPI, params.Encode())

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed with status: %s", resp.Status)
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

//go:embed static/connector/*
var connectorStaticDir embed.FS

const (
	connectorDirName      = "connector"
	connectorTemplatePath = "static/" + connectorDirName
	packageJsonName       = "package.json"
	connectorSpecName     = "connector-spec.json"
)

// newConnInitCmd is a connectors subcommand used to initialize a new connector project.
// It accepts one argument, project name, and generates appropriate directories and files
// to set up a working, sample project.
func newConnInitCommand() *cobra.Command {
	var colab bool
	var repoUrl string
	cmd := &cobra.Command{
		Use:     "init <connector-name>",
		Short:   "Initialize new connector project",
		Long:    `init sets up a new TypeScript project with sample connector included for reference.`,
		Example: "sail connectors init \"My Connector\"",
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

			if colab {
				repos, err := fetchRepositories("sailpoint-oss", "colab-saas-conn")
				if err != nil {
					fmt.Println("Error fetching repositories:", err)
					return
				}

				if len(repos) == 0 {
					fmt.Println("No repositories found.")
					return
				}

				sort.Slice(repos, func(i, j int) bool {
					return repos[i].Name < repos[j].Name
				})

				repoUrl, err = SelectColabRepo(repos)

				if err != nil {
					fmt.Println("Error selecting repository:", err)
					return
				}

				fmt.Println("Selected repository:", repoUrl)
				if err := cloneRepo(repoUrl, projName); err != nil {
					fmt.Println("Error cloning repository:", err)
				} else {
					fmt.Println("Repository cloned successfully!")
				}

			} else {
				err := fs.WalkDir(connectorStaticDir, connectorTemplatePath, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}

					if d.Name() == connectorDirName {
						return nil
					}

					if d.IsDir() {
						if err := createDir(filepath.Join(projName, d.Name())); err != nil {
							return err
						}
					} else {
						fileName := filepath.Join(projName, strings.TrimPrefix(path, connectorTemplatePath))

						data, err := connectorStaticDir.ReadFile(path)
						if err != nil {
							return err
						}

						if err := createFile(fileName, data); err != nil {
							return err
						}
					}

					if d.Name() == packageJsonName || d.Name() == connectorSpecName {
						fileAbsPath, err := filepath.Abs(filepath.Join(projName, strings.TrimPrefix(path, connectorTemplatePath)))
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
			}

			printDir(cmd.OutOrStdout(), projName, 0)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully created project '%s'.\nRun `npm install` to install dependencies.\n", projName)
		},
	}

	cmd.Flags().BoolVarP(&colab, "from-colab", "c", false, "Initialize a new connector project from a Colab connector")

	return cmd
}

func SelectColabRepo[T Repo](repos []Repo) (string, error) {
	var prompts []tui.Choice
	for i := 0; i < len(repos); i++ {
		temp := repos[i]

		prompts = append(prompts, tui.Choice{Title: temp.Name, Description: temp.Description, Id: temp.CloneURL})
	}

	intermediate, err := tui.PromptList(prompts, "Select a Colab connector to initialize")
	if err != nil {
		return "", err
	}
	return intermediate.Id, nil

}

func cloneRepo(repoURL, destination string) error {
	_, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL: repoURL,
	})
	return err
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
