/*
IdentityNow V3 API

Use these APIs to interact with the IdentityNow platform to achieve repeatable, automated processes with greater scalability. We encourage you to join the SailPoint Developer Community forum at https://developer.sailpoint.com/discuss to connect with other developers using our APIs.

API version: 3.0.0
*/

package sailpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/viper"
)

type PatConfig struct {
	ClientID     string    `mapstructure:"clientid"`
	ClientSecret string    `mapstructure:"clientsecret"`
	AccessToken  string    `mapstructure:"accesstoken"`
	Expiry       time.Time `mapstructure:"expiry"`
}

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

type OrgConfig struct {

	//Standard Variables
	Debug             bool                   `mapstructure:"debug"`
	AuthType          string                 `mapstructure:"authtype"`
	ActiveEnvironment string                 `mapstructure:"activeenvironment"`
	Environments      map[string]Environment `mapstructure:"environments"`
}

type ClientConfiguration struct {
	ClientId     string
	ClientSecret string
	BaseURL      string
	TokenURL     string
	Token        string
}

// ServerVariable stores the information about a server variable
type ServerVariable struct {
	Description  string
	DefaultValue string
	EnumValues   []string
}

// Configuration stores the configuration of the API client
type Configuration struct {
	Host                string            `json:"host,omitempty"`
	Scheme              string            `json:"scheme,omitempty"`
	DefaultHeader       map[string]string `json:"defaultHeader,omitempty"`
	UserAgent           string            `json:"userAgent,omitempty"`
	Debug               bool              `json:"debug,omitempty"`
	HTTPClient          *retryablehttp.Client
	ClientConfiguration ClientConfiguration
}

// NewConfiguration returns a new Configuration object
func NewConfiguration(clientConfiguration ClientConfiguration) *Configuration {
	cfg := &Configuration{
		DefaultHeader:       make(map[string]string),
		UserAgent:           "OpenAPI-Generator/0.1.0/go",
		Debug:               false,
		ClientConfiguration: clientConfiguration,
	}
	return cfg
}

func NewDefaultConfiguration() *Configuration {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("unable to find home directory: %s \n", err))
	}
	viper.AddConfigPath(filepath.Join(home, ".sailpoint"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("sail")

	viper.AutomaticEnv()

	if err2 := viper.ReadInConfig(); err != nil {
		if _, ok := err2.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// IGNORE they may be using env vars
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("unable to read config: %s \n", err2))
		}
	}

	var config OrgConfig
	var simpleConfig ClientConfiguration

	err3 := viper.Unmarshal(&config)

	if err3 != nil {
		panic(fmt.Errorf("unable to decode Config: %s \n", err3))
	}

	simpleConfig.BaseURL = config.Environments[config.ActiveEnvironment].BaseURL
	simpleConfig.ClientId = config.Environments[config.ActiveEnvironment].Pat.ClientID
	simpleConfig.ClientSecret = config.Environments[config.ActiveEnvironment].Pat.ClientSecret

	if os.Getenv("SAIL_BASE_URL") != "" {
		simpleConfig.BaseURL = os.Getenv("SAIL_BASE_URL")
	}
	if os.Getenv("SAIL_CLIENT_ID") != "" {
		simpleConfig.ClientId = os.Getenv("SAIL_CLIENT_ID")
	}
	if os.Getenv("SAIL_CLIENT_SECRET") != "" {
		simpleConfig.ClientSecret = os.Getenv("SAIL_CLIENT_SECRET")
	}
	simpleConfig.TokenURL = simpleConfig.BaseURL + "/oauth/token"

	return NewConfiguration(simpleConfig)
}
