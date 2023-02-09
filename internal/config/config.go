package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ErrAccessTokenExpired = fmt.Errorf("accesstoken is expired")

const (
	configFolder   = ".sailpoint"
	configYamlFile = "config.yaml"
)

type Token struct {
	AccessToken string    `mapstructure:"accesstoken"`
	Expiry      time.Time `mapstructure:"expiry"`
}

type Environment struct {
	TenantURL string    `mapstructure:"tenanturl"`
	BaseURL   string    `mapstructure:"baseurl"`
	Pat       PatConfig `mapstructure:"pat"`
	OAuth     Token     `mapstructure:"oauth"`
}

type CLIConfig struct {
	CustomExportTemplatesPath string `mapstructure:"customexporttemplatespath"`
	CustomSearchTemplatesPath string `mapstructure:"customsearchtemplatespath"`
	Debug                     bool   `mapstructure:"debug"`
	AuthType                  string `mapstructure:"authtype"`
	ActiveEnvironment         string `mapstructure:"activeenvironment"`
	Environments              map[string]Environment
}

func GetCustomSearchTemplatePath() string {
	return viper.GetString("customsearchtemplatespath")
}

func GetCustomExportTemplatePath() string {
	return viper.GetString("customexporttemplatespath")
}

func SetCustomSearchTemplatePath(customsearchtemplatespath string) {
	viper.Set("customsearchtemplatespath", customsearchtemplatespath)
}

func SetCustomExportTemplatePath(customsearchtemplatespath string) {
	viper.Set("customexporttemplatespath", customsearchtemplatespath)
}

func GetEnvironments() map[string]interface{} {
	return viper.GetStringMap("environments")
}

func GetAuthType() string {
	return strings.ToLower(viper.GetString("authtype"))
}

func SetAuthType(AuthType string) {
	viper.Set("authtype", strings.ToLower(AuthType))
}

func GetDebug() bool {
	return viper.GetBool("debug")
}

func SetDebug(Debug bool) {
	viper.Set("debug", Debug)
}

func GetActiveEnvironment() string {
	return strings.ToLower(viper.GetString("activeenvironment"))
}

func SetActiveEnvironment(activeEnv string) {
	viper.Set("activeenvironment", strings.ToLower(activeEnv))
}

func InitConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(filepath.Join(home, ".sailpoint"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("sail")

	viper.SetDefault("activeenvironment", "")
	viper.SetDefault("customexporttemplatespath", "")
	viper.SetDefault("customsearchtemplatespath", "")
	viper.SetDefault("debug", false)
	viper.SetDefault("authtype", "oauth")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// IGNORE they may be using env vars
		} else {
			// Config file was found but another error was produced
			cobra.CheckErr(err)
		}
	}
}

func InitAPIClient() *sailpoint.APIClient {
	var DevNull types.DevNull
	token, err := GetAuthToken()
	if err != nil && GetDebug() {
		color.Yellow("unable to retrieve accesstoken: %s ", err)
	}

	configuration := sailpoint.NewConfiguration(sailpoint.ClientConfiguration{Token: token, BaseURL: GetBaseUrl()})
	apiClient := sailpoint.NewAPIClient(configuration)
	apiClient.V3.GetConfig().HTTPClient.Logger = DevNull
	apiClient.Beta.GetConfig().HTTPClient.Logger = DevNull

	return apiClient
}

func GetAuthToken() (string, error) {
	err := Validate()
	if err != nil {
		return "", err
	}

	switch GetAuthType() {
	case "pat":
		if GetPatTokenExpiry().After(time.Now()) {
			return GetPatToken(), nil
		} else {
			err = PATLogin()
			if err != nil {
				return "", err
			}

			return GetPatToken(), nil
		}
	case "oauth":
		if GetOAuthTokenExpiry().After(time.Now()) {
			return GetOAuthToken(), nil
		} else {
			err = OAuthLogin()
			if err != nil {
				return "", err
			}

			return GetOAuthToken(), nil
		}
	default:
		return "", fmt.Errorf("invalid authtype configured")
	}
}

func GetBaseUrl() string {
	return viper.GetString(fmt.Sprintf("environments.%s.baseurl", GetActiveEnvironment()))
}

func GetTenantUrl() string {
	return viper.GetString(fmt.Sprintf("environments.%s.tenanturl", GetActiveEnvironment()))
}

func SetBaseUrl(baseUrl string) {
	viper.Set(fmt.Sprintf("environments.%s.baseurl", GetActiveEnvironment()), baseUrl)
}

func SetTenantUrl(tenantUrl string) {
	viper.Set(fmt.Sprintf("environments.%s.tenanturl", GetActiveEnvironment()), tenantUrl)
}

func GetTokenUrl() string {
	return GetBaseUrl() + "/oauth/token"
}

func GetAuthorizeUrl() string {
	return GetTenantUrl() + "/oauth/authorize"
}

func GetConfig() (CLIConfig, error) {
	var Config CLIConfig
	err := viper.Unmarshal(&Config)
	if err != nil {
		return Config, err
	}
	return Config, nil
}

func SaveConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(home, configFolder)); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Join(home, configFolder), 0777)
		if err != nil {
			log.Printf("failed to create %s folder for config. %v", configFolder, err)
		}
	}

	err = viper.WriteConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.SafeWriteConfig()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func Validate() error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	if config.Environments[config.ActiveEnvironment].BaseURL == "" {
		return fmt.Errorf("environment is missing BaseURL")
	}

	if config.Environments[config.ActiveEnvironment].TenantURL == "" {
		return fmt.Errorf("environment is missing TenantURL")
	}

	switch GetAuthType() {

	case "pat":

		if config.Environments[config.ActiveEnvironment].Pat.ClientID == "" {
			return fmt.Errorf("environment is missing PAT ClientID")
		}

		if config.Environments[config.ActiveEnvironment].Pat.ClientSecret == "" {
			return fmt.Errorf("environment is missing PAT ClientSecret")
		}

		return nil

	case "oauth":

		return nil

	default:

		return fmt.Errorf("invalid authtype '%s' configured", config.AuthType)

	}
}
