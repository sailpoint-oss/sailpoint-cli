package util

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/viper"
)

func GetAuthType() string {
	return strings.ToLower(viper.GetString("authtype"))
}

func GetBaseUrl() string {
	switch GetAuthType() {
	case "oauth":
		return viper.GetString("oauth.baseurl")
	case "pat":
		return viper.GetString("pat.baseurl")
	}
	return ""
}

func GetAuthToken() string {
	switch GetAuthType() {
	case "oauth":
		return viper.GetString("oauth.token.accesstoken")
	case "pat":
		return viper.GetString("pat.token.accesstoken")
	}
	return ""
}

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
