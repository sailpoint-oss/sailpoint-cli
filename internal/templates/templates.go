package templates

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
)

func GetSearchTemplates() ([]SearchTemplate, error) {
	var searchTemplates []SearchTemplate
	var templates []SearchTemplate
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	templateFiles := []string{filepath.Join(home, ".sailpoint", "search-templates.json")}

	customTemplates := config.GetCustomSearchTemplatePath()
	if customTemplates != "" {
		templateFiles = append(templateFiles, customTemplates)
	}

	envSearchTemplates := os.Getenv("SAIL_SEARCH_TEMPLATES_PATH")
	if envSearchTemplates != "" {
		templateFiles = append(templateFiles, envSearchTemplates)
	}

	for i := 0; i < len(templateFiles); i++ {
		templateFile := templateFiles[i]

		file, err := os.OpenFile(templateFile, os.O_RDWR, 0777)
		if err != nil {
			if config.GetDebug() {
				log.Log.Error("error opening file %s", templateFile)
			}
		} else {

			raw, err := io.ReadAll(file)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(raw, &templates)
			if err != nil {
				log.Log.Error("an error occured while parsing the file: %s", templateFile)
				return nil, err
			}

			searchTemplates = append(searchTemplates, templates...)
		}
	}

	err = json.Unmarshal([]byte(builtInSearchTemplates), &templates)
	if err != nil {
		color.Red("an error occured while parsing the built in templates")
		return nil, err
	}

	searchTemplates = append(searchTemplates, templates...)

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

func GetExportTemplates() ([]ExportTemplate, error) {
	var exportTemplates []ExportTemplate
	var templates []ExportTemplate
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	templateFiles := []string{filepath.Join(home, ".sailpoint", "export-templates.json")}

	customTemplates := config.GetCustomExportTemplatePath()
	if customTemplates != "" {
		templateFiles = append(templateFiles, customTemplates)
	}

	envExportTemplates := os.Getenv("SAIL_EXPORT_TEMPLATES_PATH")
	if envExportTemplates != "" {
		templateFiles = append(templateFiles, envExportTemplates)
	}

	if len(templateFiles) > 0 {
		for i := 0; i < len(templateFiles); i++ {
			templateFile := templateFiles[i]

			file, err := os.OpenFile(templateFile, os.O_RDWR, 0777)
			if err != nil {
				if config.GetDebug() {
					log.Log.Error("error opening file %s", templateFile)
				}
			} else {

				raw, err := io.ReadAll(file)
				if err != nil {
					return nil, err
				}

				err = json.Unmarshal(raw, &templates)
				if err != nil {
					log.Log.Error("an error occured while parsing the file: %s", templateFile)
					return nil, err
				}

				exportTemplates = append(exportTemplates, templates...)
			}
		}

		err = json.Unmarshal([]byte(builtInExportTemplates), &templates)
		if err != nil {
			log.Log.Error("an error occured while parsing the built in templates")
			return nil, err
		}

		exportTemplates = append(exportTemplates, templates...)

		if len(exportTemplates) > 0 {
			for i := 0; i < len(exportTemplates); i++ {
				entry := &exportTemplates[i]
				if len(entry.Variables) > 0 {
					entry.Raw, err = json.Marshal(entry.ExportBody)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return exportTemplates, nil

}

func SelectTemplate[T Template](templates []T) (string, error) {
	var prompts []tui.Choice
	for i := 0; i < len(templates); i++ {
		temp := templates[i]

		var description string
		if temp.GetVariableCount() > 0 {
			description = temp.GetDescription() + " - Accepts Input"
		} else {
			description = temp.GetDescription()
		}
		prompts = append(prompts, tui.Choice{Title: temp.GetName(), Description: description})
	}

	intermediate, err := tui.PromptList(prompts, "Select a Template")
	if err != nil {
		return "", err
	}
	return intermediate.Title, nil

}
