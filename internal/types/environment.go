package types

import (
	"time"
)

type Environment struct {
	Pat   PatConfig   `mapstructure:"pat"`
	OAuth OAuthConfig `mapstructure:"oauth"`
}

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
