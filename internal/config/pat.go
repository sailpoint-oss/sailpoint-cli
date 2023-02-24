package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type PatConfig struct {
	ClientID     string    `mapstructure:"clientid"`
	ClientSecret string    `mapstructure:"clientsecret"`
	AccessToken  string    `mapstructure:"accesstoken"`
	Expiry       time.Time `mapstructure:"expiry"`
}

func GetPatToken() string {
	return viper.GetString(fmt.Sprintf("environments.%s.pat.accesstoken", GetActiveEnvironment()))
}

func SetPatToken(token string) {
	viper.Set(fmt.Sprintf("environments.%s.pat.accesstoken", GetActiveEnvironment()), token)
}

func GetPatTokenExpiry() time.Time {
	return viper.GetTime(fmt.Sprintf("environments.%s.pat.expiry", GetActiveEnvironment()))
}

func SetPatTokenExpiry(expiry time.Time) {
	viper.Set(fmt.Sprintf("environments.%s.pat.expiry", GetActiveEnvironment()), expiry)
}

func GetPatClientID() string {
	envSecret := os.Getenv("SAIL_CLIENT_ID")
	if envSecret != "" {
		return envSecret
	} else {
		return viper.GetString(fmt.Sprintf("environments.%s.pat.clientid", GetActiveEnvironment()))
	}
}

func GetPatClientSecret() string {
	envSecret := os.Getenv("SAIL_CLIENT_SECRET")
	if envSecret != "" {
		return envSecret
	} else {
		return viper.GetString(fmt.Sprintf("environments.%s.pat.clientsecret", GetActiveEnvironment()))
	}
}

func SetPatClientID(ClientID string) {
	viper.Set(fmt.Sprintf("environments.%s.pat.clientid", GetActiveEnvironment()), ClientID)
}

func SetPatClientSecret(ClientSecret string) {
	viper.Set(fmt.Sprintf("environments.%s.pat.clientsecret", GetActiveEnvironment()), ClientSecret)
}

func PATLogin() error {
	config, err := GetConfig()
	if err != nil {
		return err
	}
	uri, err := url.Parse(GetTokenUrl())
	if err != nil {
		return err
	}

	query := &url.Values{}
	query.Add("grant_type", "client_credentials")
	uri.RawQuery = query.Encode()

	data := &url.Values{}
	data.Add("client_id", config.Environments[config.ActiveEnvironment].Pat.ClientID)
	data.Add("client_secret", config.Environments[config.ActiveEnvironment].Pat.ClientSecret)

	ctx := context.TODO()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to retrieve access token. status %s", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tResponse TokenResponse

	err = json.Unmarshal(raw, &tResponse)
	if err != nil {
		return err
	}

	now := time.Now()

	SetPatToken(tResponse.AccessToken)
	SetPatTokenExpiry(now.Add(time.Second * time.Duration(tResponse.ExpiresIn)))

	return nil
}
