package env

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new environment",
		Long:  "\nCreate a new CLI environment with tenant configuration and authentication.\n\nThis interactive command walks you through setting up a tenant URL,\nchoosing an authentication method (PAT or OAuth), and configuring credentials.\n\n",
		Example: `  sail env create
  sail env create production`,
		Aliases: []string{"c"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createOrUpdateEnv(args, false)
		},
	}
}

func createOrUpdateEnv(args []string, update bool) error {
	environments := config.GetEnvironments()

	var envName string
	if len(args) > 0 {
		envName = args[0]
	}

	if !update && envName != "" && environments[envName] != nil {
		fmt.Printf("Environment '%s' already exists. Use 'sail env update %s' to update it.\n", envName, envName)
		return nil
	}

	if update {
		fmt.Print("This utility will walk you through updating an environment.\n\n")
	} else {
		fmt.Print("This utility will walk you through creating a new environment.\n\n")
	}
	fmt.Print("Press ^C at any time to quit.\n\n")

	// Prompt for tenant name
	defaultTenant := envName
	if update && envName == "" {
		defaultTenant = config.GetActiveEnvironment()
	}

	tenant, err := tui.Input("Tenant (e.g. acme)", defaultTenant)
	if err != nil {
		return err
	}
	if tenant == "" {
		tenant = defaultTenant
	}

	if tenant == "" {
		return fmt.Errorf("tenant name is required")
	}

	// Check for existing env when creating
	if !update && environments[tenant] != nil && envName == "" {
		fmt.Printf("Environment '%s' already exists. Use 'sail env update %s' to update it.\n", tenant, tenant)
		return nil
	}

	// Determine the environment name to use in config
	effectiveName := envName
	if effectiveName == "" {
		effectiveName = tenant
	}

	tenantURL := "https://" + tenant + ".identitynow.com"
	baseURL := "https://" + tenant + ".api.identitynow.com"

	fmt.Print("\nIf the generated URLs below are correct, press Enter to accept them.\n\n")
	tenantURL, err = tui.Input("Tenant URL", tenantURL)
	if err != nil {
		return err
	}
	baseURL, err = tui.Input("Base URL", baseURL)
	if err != nil {
		return err
	}

	// Prompt for auth type
	authType, err := promptAuthType()
	if err != nil {
		return err
	}

	// Set the environment as active and configure URLs
	config.SetActiveEnvironment(effectiveName)
	config.SetTenantUrl(tenantURL)
	config.SetBaseUrl(baseURL)
	config.SetAuthType(authType)

	// Configure auth credentials inline
	switch authType {
	case "pat":
		if err := configurePAT(effectiveName); err != nil {
			return err
		}
	case "oauth":
		if err := configureOAuth(effectiveName, baseURL); err != nil {
			return err
		}
	}

	action := "created"
	if update {
		action = "updated"
	}
	fmt.Printf("\nEnvironment '%s' %s and set as active.\n", effectiveName, action)

	return nil
}

func promptAuthType() (string, error) {
	items := []tui.Choice{
		{Title: "PAT", Description: "Personal Access Token - authenticate with Client ID and Client Secret"},
		{Title: "OAuth", Description: "OAuth2.0 - sign in via the Identity Security Cloud web portal"},
	}

	choice, err := tui.PromptList(items, "Choose an authentication method")
	if err != nil {
		return "", err
	}

	return strings.ToLower(choice.Title), nil
}

func configurePAT(envName string) error {
	clientID, err := auth.PromptForClientID()
	if err != nil {
		return err
	}

	clientSecret, err := auth.PromptForClientSecret()
	if err != nil {
		return err
	}

	if err := auth.SetPatClientID(envName, clientID); err != nil {
		return err
	}
	if err := auth.SetPatClientSecret(envName, clientSecret); err != nil {
		return err
	}
	if err := auth.ResetCachePAT(envName); err != nil {
		return err
	}

	// Verify credentials by fetching a token
	fmt.Print("\nVerifying credentials... ")
	tokenURL := config.GetBaseUrl() + "/oauth/token"
	set, err := auth.PATLogin(tokenURL, clientID, clientSecret)
	if err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("credential verification failed: %w", err)
	}
	fmt.Println("OK")

	if err := auth.CachePAT(envName, set); err != nil {
		log.Warn("Failed to cache token", "error", err)
	}

	claims, err := auth.GetTokenClaims(set.AccessToken)
	if err == nil && claims["user_name"] != nil {
		fmt.Printf("Authenticated as: %v (org: %v)\n", claims["user_name"], claims["org"])
	}

	return nil
}

func configureOAuth(envName, baseURL string) error {
	fmt.Println("\nInitiating OAuth login...")
	set, err := auth.OAuthLogin(baseURL)
	if err != nil {
		return err
	}

	if set.BaseURL != "" {
		config.SetBaseUrl(set.BaseURL)
	}

	if err := auth.CacheOAuth(envName, set); err != nil {
		log.Warn("Failed to cache OAuth tokens", "error", err)
	}

	claims, err := auth.GetTokenClaims(set.AccessToken)
	if err == nil && claims["user_name"] != nil {
		fmt.Printf("Authenticated as: %v (org: %v)\n", claims["user_name"], claims["org"])
	}

	return nil
}
