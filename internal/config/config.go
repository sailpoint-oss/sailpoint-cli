package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	sailpoint "github.com/sailpoint-oss/golang-sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/viper"
	"gopkg.in/square/go-jose.v2/jwt"
)

var ErrAccessTokenExpired = fmt.Errorf("accesstoken is expired")

const (
	configFolder   = ".sailpoint"
	configYamlFile = "config.yaml"
)

type Token struct {
	AccessToken string    `mapstructure:"accesstoken"`
	Expiry      time.Time `mapstructure:"expiry"`

	RefreshToken  string    `mapstructure:"refreshtoken"`
	RefreshExpiry time.Time `mapstructure:"refreshexpiry"`
}

type Environment struct {
	TenantURL string    `mapstructure:"tenanturl"`
	BaseURL   string    `mapstructure:"baseurl"`
	Pat       PatConfig `mapstructure:"pat"`
	OAuth     Token     `mapstructure:"oauth"`
}

type CLIConfig struct {

	//Standard Variables
	ExportTemplatesPath string `mapstructure:"exporttemplatespath"`
	SearchTemplatesPath string `mapstructure:"searchtemplatespath"`
	ReportTemplatesPath string `mapstructure:"reporttemplatespath"`
	// TemplatesPath       string                 `mapstructure:"templatespath"`

	Debug             bool                   `mapstructure:"debug"`
	AuthType          string                 `mapstructure:"authtype"`
	ActiveEnvironment string                 `mapstructure:"activeenvironment"`
	Environments      map[string]Environment `mapstructure:"environments"`

	//Pipline Variables
	ClientID     string    `mapstructure:"clientid, omitempty"`
	ClientSecret string    `mapstructure:"clientsecret, omitempty"`
	BaseURL      string    `mapstructure:"base_url, omitempty"`
	AccessToken  string    `mapstructure:"accesstoken"`
	Expiry       time.Time `mapstructure:"expiry"`
}

func GetCustomSearchTemplatePath() string {
	return viper.GetString("searchtemplatespath")
}

func GetCustomExportTemplatePath() string {
	return viper.GetString("exporttemplatespath")
}

func GetCustomReportTemplatePath() string {
	return viper.GetString("reporttemplatespath")
}

func SetCustomSearchTemplatePath(customsearchtemplatespath string) {
	viper.Set("searchtemplatespath", customsearchtemplatespath)
}

func SetCustomExportTemplatePath(customsearchtemplatespath string) {
	viper.Set("exporttemplatespath", customsearchtemplatespath)
}

func SetCustomReportTemplatePath(customreporttemplatespath string) {
	viper.Set("reporttemplatespath", customreporttemplatespath)
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

func InitConfig() error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	viper.AddConfigPath(filepath.Join(home, ".sailpoint"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("sail")

	viper.SetDefault("authtype", "pat")
	viper.SetDefault("exporttemplatespath", "")
	viper.SetDefault("searchtemplatespath", "")
	viper.SetDefault("reporttemplatespath", "")
	viper.SetDefault("debug", false)
	viper.SetDefault("activeenvironment", "default")

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

	if GetDebug() {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	}

	return nil
}

func InitAPIClient() (*sailpoint.APIClient, error) {
	var apiClient *sailpoint.APIClient

	err := Validate()
	if err != nil {
		return apiClient, err
	}

	token, err := GetAuthToken()
	if err != nil {
		log.Debug("unable to retrieve accesstoken", "error", err)
	}

	configuration := sailpoint.NewConfiguration(sailpoint.ClientConfiguration{Token: token, BaseURL: GetBaseUrl()})
	apiClient = sailpoint.NewAPIClient(configuration)
	if GetDebug() {
		logger := log.NewWithOptions(os.Stdout, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			Level:           log.DebugLevel,
		})
		debugLogger := logger.StandardLog(log.StandardLogOptions{ForceLevel: log.DebugLevel})
		apiClient.V3.GetConfig().HTTPClient.Logger = debugLogger
		apiClient.Beta.GetConfig().HTTPClient.Logger = debugLogger
	} else {
		var DevNull types.DevNull
		apiClient.V3.GetConfig().HTTPClient.Logger = DevNull
		apiClient.Beta.GetConfig().HTTPClient.Logger = DevNull
	}

	return apiClient, nil
}

func CheckToken(tokenString string) error {
	var claims map[string]interface{}

	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return err
	}

	token.UnsafeClaimsWithoutVerification(&claims)

	if claims["user_name"] == nil {
		log.Warn("It looks like the token you are using is missing a user context, this will cause many of the CLI commands to fail.")
	}

	log.Debug("Token Debug Info", "user_name", claims["user_name"], "org", claims["org"], "pod", claims["pod"])

	return nil
}

func SetTime(inputTime time.Time) string {
	return inputTime.Format(time.RFC3339)
}

func GetTime(inputString string) (time.Time, error) {
	var outputTime time.Time
	outputTime, err := time.Parse(time.RFC3339, inputString)
	if err != nil {
		return outputTime, err
	}
	return outputTime, nil
}

func GetAuthToken() (string, error) {

	var token string

	err := InitConfig()
	if err != nil {
		return "", err
	}

	err = Validate()
	if err != nil {
		return "", err
	}

	switch GetAuthType() {

	case "pat":

		authExpiry, _ := GetPatTokenExpiry()

		if authExpiry.After(time.Now()) {

			tempToken, err := GetPatToken()
			if err != nil {
				return token, err
			}

			token = tempToken

		} else {

			err = PATLogin()
			if err != nil {
				return "", err
			}

			tempToken, err := GetPatToken()
			if err != nil {
				return token, err
			}

			token = tempToken
		}

	case "oauth":

		authExpiry, _ := GetOAuthTokenExpiry()
		refreshExpiry, _ := GetOAuthRefreshExpiry()

		if authExpiry.After(time.Now()) {

			tempToken, err := GetOAuthToken()
			if err != nil {
				return token, err
			}

			token = tempToken

		} else if refreshExpiry.After(time.Now()) {

			err := RefreshOAuth()
			if err != nil {
				return token, err
			}

			tempToken, err := GetOAuthToken()
			if err != nil {
				return token, err
			}

			token = tempToken

		} else {

			err = OAuthLogin()
			if err != nil {
				return "", err
			}

			tempToken, err := GetOAuthToken()
			if err != nil {
				return token, err
			}

			token = tempToken

		}

	default:
		return "", fmt.Errorf("invalid authtype configured")
	}

	err = CheckToken(token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func GetBaseUrl() string {
	envBaseUrl := os.Getenv("SAIL_BASE_URL")
	if envBaseUrl != "" {
		return envBaseUrl
	} else {
		return viper.GetString("environments." + GetActiveEnvironment() + ".baseurl")
	}
}

func GetTenantUrl() string {
	return viper.GetString("environments." + GetActiveEnvironment() + ".tenanturl")
}

func SetBaseUrl(baseUrl string) {
	viper.Set("environments."+GetActiveEnvironment()+".baseurl", baseUrl)
}

func SetTenantUrl(tenantUrl string) {
	viper.Set("environments."+GetActiveEnvironment()+".tenanturl", tenantUrl)
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
			log.Warn("failed to create %s folder for config. %v", configFolder, err)
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
	var errors int
	authType := GetAuthType()

	switch authType {

	case "pat":

		if GetBaseUrl() == "" {
			log.Error("configured environment is missing BaseURL")
			errors++
		}

		patClientID, err := GetPatClientID()
		if err != nil {
			return err
		}
		patClientSecret, err := GetPatClientSecret()
		if err != nil {
			return err
		}

		if patClientID == "" {
			log.Error("configured environment is missing PAT ClientID")
			errors++
		}

		if patClientSecret == "" {
			log.Error("configured environment is missing PAT ClientSecret")
			errors++
		}

	case "oauth":

		if GetBaseUrl() == "" {
			log.Error("configured environment is missing BaseURL")
			errors++
		}

		if GetTenantUrl() == "" {
			log.Error("configured environment is missing TenantURL")
			errors++
		}

	default:

		log.Error("invalid authtype '%s' configured", authType)
		errors++

	}

	if errors > 0 {
		return fmt.Errorf("configuration invalid, errors: %v", errors)
	}

	return nil
}
