package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

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

func (c OrgConfig) Validate() error {
	debug := viper.GetBool("debug")
	switch c.AuthType {
	case "PAT":
		if c.Pat.TokenUrl == "" {
			return fmt.Errorf("missing PAT TokenURL configuration value")
		}
		if c.Pat.ClientID == "" {
			return fmt.Errorf("missing PAT ClientID configuration value")
		}
		if c.Pat.ClientSecret == "" {
			return fmt.Errorf("missing PAT ClientSecret configuration value")
		}
		return nil
	case "OAuth":
		if c.OAuth.AuthUrl == "" {
			return fmt.Errorf("missing OAuth URL configuration value")
		}
		if c.OAuth.ClientID == "" {
			return fmt.Errorf("missing OAuth ClientID configuration value")
		}
		if c.OAuth.ClientSecret == "" && debug {
			color.Yellow("missing OAuth ClientSecret configuration value")
		}
		if c.OAuth.Redirect.Path == "" {
			return fmt.Errorf("missing OAuth Redirect Path configuration value")
		}
		if c.OAuth.Redirect.Port == 0 {
			return fmt.Errorf("missing OAuth Redirect Port configuration value")
		}
		if c.OAuth.TokenUrl == "" {
			return fmt.Errorf("missing OAuth TokenUrl configuration value")
		}
		return nil
	default:
		return errors.New("configured authtype is invalid or missing")
	}

}
