package util

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/log"
	"github.com/mrz1836/go-sanitize"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/search"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/viper"
)

var renderer *glamour.TermRenderer

func init() {
	var err error
	renderer, err = glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
	)
	if err != nil {
		log.Warn("Markdown renderer unavailable; falling back to plain help text", "error", err)
	}

}

func PrettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Error("Error marshalling interface", "error", err)
	}
	return (string(b))
}

func SanitizeFileName(fileName string) string {
	return sanitize.PathName(fileName)
}

func RenderMarkdown(markdown string) string {
	if renderer == nil {
		return markdown
	}
	out, err := renderer.Render(markdown)
	if err != nil {
		log.Warn("Failed to render markdown; falling back to plain text", "error", err)
		return markdown
	}

	return out
}

type Help struct {
	Long    string
	Example string
}

func ParseHelp(help string) Help {
	helpParser, err := regexp.Compile(`==([A-Za-z]+)==([\s\S]*?)====`)
	if err != nil {
		log.Warn("Failed to compile help parser", "error", err)
		return Help{}
	}

	matches := helpParser.FindAllStringSubmatch(help, -1)

	var helpObj Help
	for _, set := range matches {
		switch strings.ToLower(set[1]) {
		case "long":
			helpObj.Long = RenderMarkdown(set[2])
		case "example":
			helpObj.Example = RenderMarkdown(set[2])
		}
	}

	return helpObj
}

func getTextBetween(url, start, end string) string {
	startIndex := strings.Index(url, start)
	if startIndex == -1 {
		return ""
	}
	endIndex := strings.Index(url[startIndex+len(start):], end)
	if endIndex == -1 {
		return ""
	}
	return url[startIndex+len(start) : startIndex+len(start)+endIndex]
}

func CreateOrUpdateEnvironment(environmentName string, update bool) error {
	environments := config.GetEnvironments()

	if environments[environmentName] != nil && !update {
		fmt.Print("Environment already exists\n\n To update the environment use `sail env update`.\n\n")
		return nil
	} else {
		if update {
			fmt.Print("This utility will walk you through updating an existing environment.\n\n")

		} else {
			fmt.Print("This utility will walk you through creating a new environment.\n\n")
		}

		fmt.Print("Press ^C at any time to quit.\n\n")

		tenant := ""

		var defaultTenant string
		if update && environmentName == "" {
			defaultTenant = config.GetActiveEnvironment()
		} else if update {
			defaultTenant = getTextBetween(viper.GetString("environments."+environmentName+".tenanturl"), "//", ".")
		} else {
			defaultTenant = environmentName
		}
		var err error
		tenant, err = tui.Input("Tenant Name (e.g. acme)", defaultTenant)
		if err != nil {
			return err
		}

		if !update {
			if environments[tenant] != nil {
				fmt.Print("Environment already exists\n\n To update the environment use `sail env update `" + tenant + ".\n\n")
				return nil
			}
		}

		if tenant == "" {
			tenant = environmentName
		}

		tenantUrl := "https://" + tenant + ".identitynow.com"
		baseUrl := "https://" + tenant + ".api.identitynow.com"

		fmt.Print("\nIf the generated URLs are correct, press Enter to accept them.\n\n")
		tenantUrl, err = tui.Input("Tenant URL", tenantUrl)
		if err != nil {
			return err
		}
		baseUrl, err = tui.Input("Base URL", baseUrl)
		if err != nil {
			return err
		}

		authType, err := tui.Input("Authentication Type (oauth, pat)", "")
		if err != nil {
			return err
		}

		if authType == "pat" {

			clientID, err := config.PromptForClientID()
			if err != nil {
				return err
			}

			ClientSecret, err := config.PromptForClientSecret()
			if err != nil {
				return err
			}

			if environmentName != "" {
				config.SetActiveEnvironment(environmentName)
			} else {
				config.SetActiveEnvironment(tenant)
			}

			err = config.SetPatClientSecret(ClientSecret)
			if err != nil {
				return err
			}

			err = config.ResetCachePAT()
			if err != nil {
				return err
			}

			config.SetTenantUrl(tenantUrl)
			config.SetBaseUrl(baseUrl)
			config.SetAuthType(authType)
			config.SetPatClientID(clientID)
		}

		if authType == "oauth" {

			if environmentName != "" {
				config.SetActiveEnvironment(environmentName)
			} else {
				config.SetActiveEnvironment(tenant)
			}

			config.SetTenantUrl(tenantUrl)
			config.SetBaseUrl(baseUrl)
			config.SetAuthType(authType)
			config.GetAuthToken()
		}

		fmt.Print("\n\nEnvironment successfully created.\n\n")
		fmt.Print("You can change your authentication type at any time by running `sail set auth`.\n\n")

		if authType == "pat" {
			fmt.Print("You can change your client id and secret at any time by running `sail set pat`.\n\n")
		}
	}
	return nil
}

func SelectIdentity[T search.Identity](identities []search.Identity) (string, error) {
	var prompts []tui.Choice
	for i := 0; i < len(identities); i++ {
		temp := identities[i]

		prompts = append(prompts, tui.Choice{Title: temp.DisplayName, Description: temp.Email, Id: temp.ID})
	}

	intermediate, err := tui.PromptList(prompts, "Select an Identity to preview the transform")
	if err != nil {
		return "", err
	}
	return intermediate.Id, nil

}
