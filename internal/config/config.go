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

	//Standard Variables
	CustomExportTemplatesPath string                 `mapstructure:"customexporttemplatespath"`
	CustomSearchTemplatesPath string                 `mapstructure:"customsearchtemplatespath"`
	Debug                     bool                   `mapstructure:"debug"`
	AuthType                  string                 `mapstructure:"authtype"`
	ActiveEnvironment         string                 `mapstructure:"activeenvironment"`
	Environments              map[string]Environment `mapstructure:"environments"`

	//Pipline Variables
	ClientID     string    `mapstructure:"clientid, omitempty"`
	ClientSecret string    `mapstructure:"clientsecret, omitempty"`
	BaseURL      string    `mapstructure:"base_url, omitempty"`
	AccessToken  string    `mapstructure:"accesstoken"`
	Expiry       time.Time `mapstructure:"expiry"`
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
	configAuthType := strings.ToLower(viper.GetString("authtype"))
	envAuthType := strings.ToLower(os.Getenv("SAIL_AUTH_TYPE"))
	if envAuthType == "pipeline" {
		return envAuthType
	} else {
		return configAuthType
	}
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

func InitConfig() error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	viper.AddConfigPath(filepath.Join(home, ".sailpoint"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("sail")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// IGNORE they may be using env vars
		} else {
			// Config file was found but another error was produced
			return err
		}
	}

	err = Validate()
	if err != nil {
		return err
	}

	return nil
}

func InitAPIClient() *sailpoint.APIClient {
	token, err := GetAuthToken()
	if err != nil && GetDebug() {
		color.Yellow("unable to retrieve accesstoken: %s ", err)
	}

	configuration := sailpoint.NewConfiguration(sailpoint.ClientConfiguration{Token: token, BaseURL: GetBaseUrl()})
	apiClient := sailpoint.NewAPIClient(configuration)
	if !GetDebug() {
		var DevNull types.DevNull
		apiClient.V3.GetConfig().HTTPClient.Logger = DevNull
		apiClient.Beta.GetConfig().HTTPClient.Logger = DevNull
	}

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
		return "", fmt.Errorf("oauth is not currently supported")
		// if GetOAuthTokenExpiry().After(time.Now()) {
		// 	return GetOAuthToken(), nil
		// } else {
		// 	err = OAuthLogin()
		// 	if err != nil {
		// 		return "", err
		// 	}

		// 	return GetOAuthToken(), nil
		// }
	case "pipeline":
		if GetPipelineTokenExpiry().After(time.Now()) {
			return GetPipelineToken(), nil
		} else {
			err = PipelineLogin()
			if err != nil {
				return "", err
			}

			return GetPipelineToken(), nil
		}
	default:
		return "", fmt.Errorf("invalid authtype configured")
	}
}

func GetBaseUrl() string {
	configBaseUrl := viper.GetString(fmt.Sprintf("environments.%s.baseurl", GetActiveEnvironment()))
	envBaseUrl := os.Getenv("SAIL_BASE_URL")
	if envBaseUrl != "" {
		return envBaseUrl
	} else {
		return configBaseUrl
	}
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
	authType := GetAuthType()

	switch authType {

	case "pat":

		if config.Environments[config.ActiveEnvironment].BaseURL == "" {
			return fmt.Errorf("configured environment is missing BaseURL")
		}

		if config.Environments[config.ActiveEnvironment].Pat.ClientID == "" {
			return fmt.Errorf("configured environment is missing PAT ClientID")
		}

		if config.Environments[config.ActiveEnvironment].Pat.ClientSecret == "" {
			return fmt.Errorf("configured environment is missing PAT ClientSecret")
		}

		return nil

	case "oauth":
		return fmt.Errorf("oauth is not currently supported")

		// if config.Environments[config.ActiveEnvironment].BaseURL == "" {
		// 	return fmt.Errorf("configured environment is missing BaseURL")
		// }

		// if config.Environments[config.ActiveEnvironment].TenantURL == "" {
		// 	return fmt.Errorf("configured environment is missing TenantURL")
		// }

		// return nil

	case "pipeline":

		if os.Getenv("SAIL_BASE_URL") == "" {
			return fmt.Errorf("pipeline environment is missing SAIL_BASE_URL")
		}

		if os.Getenv("SAIL_CLIENT_ID") == "" {
			return fmt.Errorf("pipeline environment is missing SAIL_CLIENT_ID")
		}

		if os.Getenv("SAIL_CLIENT_SECRET") == "" {
			return fmt.Errorf("pipeline environment is missing SAIL_CLIENT_SECRET")
		}

		return nil

	default:

		return fmt.Errorf("invalid authtype '%s' configured", authType)

	}
}
