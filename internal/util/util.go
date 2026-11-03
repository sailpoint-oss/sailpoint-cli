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
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
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
		panic(err)
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
	out, err := renderer.Render(markdown)
	if err != nil {
		panic(err)
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
		panic(err)
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

		tenant, tenantUrl, baseUrl, err := promptForTenant(environmentName, update)
		if err != nil {
			return err
		}

		if !update {
			if environments[tenant] != nil {
				fmt.Print("Environment already exists\n\n To update the environment use `sail env update " + tenant + "`.\n\n")
				return nil
			}
		}

		tenantUrl, baseUrl = confirmTenantUrls(tenantUrl, baseUrl)

		authType := terminal.InputPrompt("Authentication Type (oauth, pat):")

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

// promptForTenant asks for a tenant name or URL and returns the environment
// name, tenant URL and API base URL. On update, an empty answer keeps the URLs
// already stored for the environment.
func promptForTenant(environmentName string, update bool) (string, string, string, error) {
	defaultName := environmentName
	if update && defaultName == "" {
		defaultName = config.GetActiveEnvironment()
	}

	storedTenantUrl, storedBaseUrl := "", ""
	if update {
		storedTenantUrl = viper.GetString("environments." + defaultName + ".tenanturl")
		storedBaseUrl = viper.GetString("environments." + defaultName + ".baseurl")
	}

	promptDefault := defaultName
	if storedTenantUrl != "" {
		promptDefault = storedTenantUrl
	}

	for attempt := 0; attempt < 3; attempt++ {
		input := terminal.InputPrompt("Tenant name or URL (ie: acme, or https://acme.identitynow-demo.com): (" + promptDefault + ")")

		if input == "" && storedTenantUrl != "" && storedBaseUrl != "" {
			return defaultName, storedTenantUrl, storedBaseUrl, nil
		}

		if input == "" {
			input = defaultName
		}

		tenant, tenantUrl, baseUrl, err := ParseTenantInput(input)
		if err != nil {
			log.Warn("Invalid tenant name or URL", "error", err)
			continue
		}

		if environmentName != "" {
			tenant = environmentName
		}

		return tenant, tenantUrl, baseUrl, nil
	}

	return "", "", "", fmt.Errorf("no valid tenant name or URL provided")
}

// confirmTenantUrls checks the API URL against the tenant. When the check
// fails, the user can correct the URLs. The URLs are kept even if the check
// still fails, so tenants that are not reachable from this network still work.
func confirmTenantUrls(tenantUrl, baseUrl string) (string, string) {
	err := config.CheckTenantAPI(baseUrl)
	if err == nil {
		fmt.Print("\n\u2714 Found tenant at " + tenantUrl + " (API: " + baseUrl + ")\n\n")
		return tenantUrl, baseUrl
	}
	log.Warn("Could not confirm the tenant API URL", "url", baseUrl, "error", err)

	fmt.Print("\nThe following two prompts will allow you to set a custom base and tenant url if the generated URL\ndoes not apply. If the generated URL is correct simply press enter to proceed\n\n")
	confirmTenantUrl := terminal.InputPrompt("Tenant URL (ie: https://{tenant}.identitynow.com): (" + tenantUrl + ")")
	confirmBaseURL := terminal.InputPrompt("Base URL (ie: https://{tenant}.api.identitynow.com): (" + baseUrl + ")")

	if confirmTenantUrl != "" {
		tenantUrl = confirmTenantUrl
	}

	if confirmBaseURL != "" {
		baseUrl = confirmBaseURL
	}

	if confirmTenantUrl != "" || confirmBaseURL != "" {
		if err := config.CheckTenantAPI(baseUrl); err != nil {
			log.Warn("Could not confirm the tenant API URL, saving it anyway", "url", baseUrl, "error", err)
		} else {
			fmt.Print("\n\u2714 Found tenant at " + tenantUrl + " (API: " + baseUrl + ")\n\n")
		}
	}

	return tenantUrl, baseUrl
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
