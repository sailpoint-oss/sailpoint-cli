package templates

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
)

func GetSearchTemplates() ([]SearchTemplate, error) {
	var searchTemplates []SearchTemplate
	var templates []SearchTemplate
	var builtInTemplates []SearchTemplate
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

			log.Debug("error opening file", "file", templateFile)

		} else {

			raw, err := io.ReadAll(file)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(raw, &templates)
			if err != nil {
				log.Error("an error occurred while parsing the file: %s", templateFile)
				return nil, err
			}

			searchTemplates = append(searchTemplates, templates...)
		}
	}

	err = json.Unmarshal([]byte(builtInSearchTemplates), &builtInTemplates)
	if err != nil {
		color.Red("an error occurred while parsing the built in templates")
		return nil, err
	}

	searchTemplates = append(searchTemplates, builtInTemplates...)

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

				log.Debug("error opening file %s", templateFile)

			} else {

				raw, err := io.ReadAll(file)
				if err != nil {
					return nil, err
				}

				err = json.Unmarshal(raw, &templates)
				if err != nil {
					log.Debug("an error occurred while parsing the file: %s", templateFile)
					return nil, err
				}

				exportTemplates = append(exportTemplates, templates...)
			}
		}

		err = json.Unmarshal([]byte(builtInExportTemplates), &templates)
		if err != nil {
			log.Error("an error occurred while parsing the built in templates")
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

func GetReportTemplates() ([]ReportTemplate, error) {
	var reportTemplates []ReportTemplate
	var templates []ReportTemplate
	var buildInTemplates []ReportTemplate
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	templateFiles := []string{filepath.Join(home, ".sailpoint", "report-templates.json")}

	customTemplates := config.GetCustomReportTemplatePath()
	if customTemplates != "" {
		templateFiles = append(templateFiles, customTemplates)
	}

	envSearchTemplates := os.Getenv("SAIL_REPORT_TEMPLATES_PATH")
	if envSearchTemplates != "" {
		templateFiles = append(templateFiles, envSearchTemplates)
	}

	for i := 0; i < len(templateFiles); i++ {
		templateFile := templateFiles[i]

		file, err := os.OpenFile(templateFile, os.O_RDWR, 0777)
		if err != nil {

			log.Debug("error opening file %s", templateFile)

		} else {

			raw, err := io.ReadAll(file)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(raw, &templates)
			if err != nil {
				log.Error("an error occured while parsing the file: %s", templateFile)
				return nil, err
			}

			reportTemplates = append(reportTemplates, templates...)
		}
	}

	err = json.Unmarshal([]byte(builtInReportTemplates), &buildInTemplates)
	if err != nil {
		color.Red("an error occured while parsing the built in templates")
		return nil, err
	}

	reportTemplates = append(reportTemplates, buildInTemplates...)

	for i := 0; i < len(reportTemplates); i++ {
		entry := &reportTemplates[i]
		if len(entry.Variables) > 0 {
			entry.Raw, err = json.Marshal(entry.Queries)
			if err != nil {
				return nil, err
			}
		}
	}
	return reportTemplates, nil
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
