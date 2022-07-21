// Copyright (c) 2022, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint/sp-cli/client"
	"github.com/sailpoint/sp-cli/validate"
	"github.com/spf13/cobra"
	"gopkg.in/alessio/shellescape.v1"
	"gopkg.in/yaml.v2"
)

type Source struct {
	// Name represents name of a source (github, smartsheet, freshservice, etc)
	Name string `yaml:"name"`
	// Repository is a link for a connector repository
	Repository string `yaml:"repository"`
	// RepositoryRef is a branch that uses for service starts
	RepositoryRef string `yaml:"repositoryRef"`
	// Config is an authentication data for service startup
	Config string `yaml:"config"`
	// ReadOnly is a flag that indicates the validation checks with data modification ('true') or without it ('false').
	ReadOnly bool `yaml:"readOnly"`
}

// ValidationResults represents validation results for every source
type ValidationResults struct {
	sourceName string
	results    map[string]*tablewriter.Table
}

const (
	connectorInstanceEndpoint = "http://localhost:3000"
	sourceFile                = "./source.yaml"
)

func (v *ValidationResults) Render() {
	fmt.Println(aurora.Blue(fmt.Sprintf("%s connectors validation results", v.sourceName)).String())
	for connectorID, result := range v.results {
		fmt.Println(aurora.Blue(fmt.Sprintf("Connector %s", connectorID)).String())
		result.Render()
		fmt.Println("---------------------------------------------------------")
	}
}

func newConnValidateSourcesCmd(apiClient client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate-sources",
		Short:   "Validate connectors behavior",
		Long:    "Validate connectors behavior from a list that stores in sources.yaml",
		Example: "sp conn validate-sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()

			listOfSources, err := getSourceFromFile(sourceFile)
			if err != nil {
				return err
			}

			var results []ValidationResults

			for _, source := range listOfSources {

				instance, tempFolder, err := runInstanceForValidation(source)
				if err != nil {
					return err
				}

				res, err := validateConnectors(ctx, apiClient, source, endpoint)
				if err != nil {
					return err
				}

				err = instance.Process.Signal(syscall.SIGTERM)
				if err != nil {
					return err
				}

				if instance.ProcessState != nil {
					return errors.New(fmt.Sprintf("%s instance wasn't stopped", source.Name))
				}

				err = os.RemoveAll(fmt.Sprintf("/%s", tempFolder))
				if err != nil {
					return err
				}

				results = append(results, *res)
			}

			for _, r := range results {
				r.Render()
			}

			return nil
		},
	}

	return cmd
}

func getSourceFromFile(filePath string) ([]Source, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config []Source

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return config, err
}

func validateConnectors(ctx context.Context, apiClient client.Client, source Source, endpoint string) (*ValidationResults, error) {
	resp, err := apiClient.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 response for getting all %s connectors: %s\nbody: %s", source.Name, resp.Status, body)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var conns []connector
	err = json.Unmarshal(raw, &conns)
	if err != nil {
		return nil, err
	}

	valRes := &ValidationResults{
		sourceName: source.Name,
		results:    make(map[string]*tablewriter.Table),
	}

	connector := conns[len(conns)-1]

	cc, err := connClientWithCustomParams(apiClient, json.RawMessage(source.Config), connector.ID, "0", connectorInstanceEndpoint)
	if err != nil {
		log.Println(err)
	}

	validator := validate.NewValidator(validate.Config{
		Check:    "",
		ReadOnly: source.ReadOnly,
	}, cc)

	results, err := validator.Run(ctx)
	if err != nil {
		log.Println(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Result", "Errors", "Warnings", "Skipped"})
	for _, res := range results {
		var result = aurora.Green("PASS")
		if len(res.Errors) > 0 {
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

	valRes.results[connector.ID] = table

	return valRes, err
}

func createTempFolder() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path, err := os.MkdirTemp(homeDir, "*")
	if err != nil {
		return "", err
	}

	return path, err
}

func runInstanceForValidation(source Source) (*exec.Cmd, string, error) {
	path, err := createTempFolder()
	if err != nil {
		return nil, "", err
	}

	cloneRepo := exec.Command("git", "clone", shellescape.Quote(source.Repository), path)

	if err := cloneRepo.Run(); err != nil {
		return nil, "", err
	}

	log.Printf("Repo for %s is cloned\n", source.Name)

	checkoutRepoRef := exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s && git checkout %s", path, shellescape.Quote(source.RepositoryRef)))
	if err := checkoutRepoRef.Run(); err != nil {
		return nil, "", err
	}

	log.Printf("git checkout to %s\n", source.RepositoryRef)

	cmd := exec.Command("npm", "install", "--prefix", path)
	if err := cmd.Run(); err != nil {
		return nil, "", err
	}

	log.Println("Npm install is finished")

	err = ExecCommand("/bin/sh", "-c", fmt.Sprintf("npm run dev --prefix %s", path))
	if err != nil {
		return nil, "", err
	}

	for {
		_, err := http.Get(connectorInstanceEndpoint)
		if err == nil {
			log.Printf("Service %s is successfully started for validation\n", source.Name)
			break
		}
		time.Sleep(time.Second * 5)
	}

	return cmd, path, err
}
