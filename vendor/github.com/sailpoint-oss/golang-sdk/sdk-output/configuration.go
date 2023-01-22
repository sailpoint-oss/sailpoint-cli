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

type Redirect struct {
	Port int    `mapstructure:"port"`
	Path string `mapstructure:"path"`
}

type Token struct {
	AccessToken string    `mapstructure:"accesstoken"`
	Expiry      time.Time `mapstructure:"expiry"`
}

type OAuthConfig struct {
	Tenant       string   `mapstructure:"tenant"`
	AuthUrl      string   `mapstructure:"authurl"`
	BaseUrl      string   `mapstructure:"baseurl"`
	TokenUrl     string   `mapstructure:"tokenurl"`
	Redirect     Redirect `mapstructure:"redirect"`
	ClientSecret string   `mapstructure:"clientSecret"`
	ClientID     string   `mapstructure:"clientid"`
	Token        Token    `mapstructure:"token"`
}

type PatConfig struct {
	Tenant       string `mapstructure:"tenant"`
	BaseUrl      string `mapstructure:"baseurl"`
	TokenUrl     string `mapstructure:"tokenurl"`
	ClientSecret string `mapstructure:"clientSecret"`
	ClientID     string `mapstructure:"clientid"`
	Token        Token  `mapstructure:"token"`
}

type OrgConfig struct {
	Pat      PatConfig   `mapstructure:"pat"`
	OAuth    OAuthConfig `mapstructure:"oauth"`
	AuthType string      `mapstructure:"authtype"`
	Debug    bool        `mapstructure:"debug"`
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
		panic(fmt.Errorf("Unable to find home directory: %s \n", err))
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
			panic(fmt.Errorf("Unable to read config: %s \n", err2))
		}
	}

	var config OrgConfig
	var simpleConfig ClientConfiguration

	err3 := viper.Unmarshal(&config)

	if err3 != nil {
		panic(fmt.Errorf("Unable to decode Config: %s \n", err3))
	}

	simpleConfig.BaseURL = config.Pat.BaseUrl
	simpleConfig.ClientId = config.Pat.ClientID
	simpleConfig.ClientSecret = config.Pat.ClientSecret
	simpleConfig.TokenURL = config.Pat.TokenUrl

	return NewConfiguration(simpleConfig)
}
