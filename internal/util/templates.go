package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/viper"
)

func GetSearchTemplates() ([]types.SearchTemplate, error) {
	var searchTemplates []types.SearchTemplate
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	templateFiles := []string{filepath.Join(home, ".sailpoint", "search-templates.json")}

	customTemplates := viper.GetString("customSearchTemplatesPath")
	if customTemplates != "" {
		templateFiles = append(templateFiles, customTemplates)
	}

	for i := 0; i < len(templateFiles); i++ {
		var templates []types.SearchTemplate
		templateFile := templateFiles[i]

		file, err := os.OpenFile(templateFile, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			return nil, err
		}

		raw, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(raw, &templates)
		if err != nil {
			color.Red("an error occured while parsing the file: %s", templateFile)
			return nil, err
		}

		searchTemplates = append(searchTemplates, templates...)
	}

	for i := 0; i < len(searchTemplates); i++ {
		entry := &searchTemplates[i]
		if len(entry.Variables) > 0 {
			entry.Raw, err = json.Marshal(entry.SearchQuery)
			if err != nil {
				return nil, err
			}
		}
	}
	return searchTemplates, nil
}

func GetExportTemplates() ([]types.ExportTemplate, error) {
	var exportTemplates []types.ExportTemplate
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	templateFiles := []string{filepath.Join(home, ".sailpoint", "export-templates.json")}

	customTemplates := viper.GetString("customExportTemplatesPath")
	if customTemplates != "" {
		templateFiles = append(templateFiles, customTemplates)
	}

	for i := 0; i < len(templateFiles); i++ {
		var templates []types.ExportTemplate
		templateFile := templateFiles[i]

		file, err := os.OpenFile(templateFile, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			return nil, err
		}

		raw, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(raw, &templates)
		if err != nil {
			color.Red("an error occured while parsing the file: %s", templateFile)
			return nil, err
		}

		exportTemplates = append(exportTemplates, templates...)
	}

	for i := 0; i < len(exportTemplates); i++ {
		entry := &exportTemplates[i]
		if len(entry.Variables) > 0 {
			entry.Raw, err = json.Marshal(entry.ExportBody)
			if err != nil {
				return nil, err
			}
		}
	}
	return exportTemplates, nil
}

func SelectTemplate[T types.Template](templates []T) (string, error) {
	var prompts []types.Choice
	for i := 0; i < len(templates); i++ {
		temp := templates[i]

		var description string
		if temp.GetVariableCount() > 0 {
			description = fmt.Sprintf("%s - Accepts Input", temp.GetDescription())
		} else {
			description = temp.GetDescription()
		}
		prompts = append(prompts, types.Choice{Title: temp.GetName(), Description: description})
	}

	intermediate, err := tui.PromptList(prompts, "Select a Template")
	if err != nil {
		return "", err
	}
	return intermediate.Title, nil

}
