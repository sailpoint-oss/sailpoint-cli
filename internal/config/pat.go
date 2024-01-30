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

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	keyring "github.com/zalando/go-keyring"
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

type PATSet struct {
	AccessToken  string
	AccessExpiry time.Time
}

func ResetCachePAT() error {

	token, err := GetPatToken()
	if token != "" && err == nil {

		err = DeletePatToken()
		if err != nil {
			return err
		}
	}

	expiry, err := GetPatTokenExpiry()
	if !expiry.IsZero() && err == nil {
		err = DeletePatTokenExpiry()
		if err != nil {
			return err
		}
	}

	return nil
}

func CachePAT(set PATSet) error {
	var err error

	err = SetPatToken(set.AccessToken)
	if err != nil {
		return err
	}

	err = SetPatTokenExpiry(set.AccessExpiry)
	if err != nil {
		return err
	}

	return nil
}

func DeletePatToken() error {
	err := keyring.Delete("environments.pat.accesstoken", GetActiveEnvironment())
	if err != nil {
		return err
	}
	return nil
}

func GetPatToken() (string, error) {
	value, err := keyring.Get("environments.pat.accesstoken", GetActiveEnvironment())
	if err != nil {
		return "", err
	}
	return value, nil
}

func SetPatToken(token string) error {
	err := keyring.Set("environments.pat.accesstoken", GetActiveEnvironment(), token)
	if err != nil {
		return err
	}
	return nil
}

func DeletePatTokenExpiry() error {
	err := keyring.Delete("environments.pat.expiry", GetActiveEnvironment())
	if err != nil {
		return err
	}
	return nil
}

func GetPatTokenExpiry() (time.Time, error) {
	valueString, err := keyring.Get("environments.pat.expiry", GetActiveEnvironment())
	if err != nil {
		return time.Time{}, err
	}

	valueTime, err := GetTime(valueString)
	if err != nil {
		return valueTime, err
	}

	return valueTime, nil
}

func SetPatTokenExpiry(expiry time.Time) error {
	err := keyring.Set("environments.pat.expiry", GetActiveEnvironment(), SetTime(expiry))
	if err != nil {
		return err
	}
	return nil
}

func GetClientID(env string) (string, error) {
	value, err := keyring.Get("environments.pat.clientid", env)
	if err != nil {
		log.Error("issue retrieving clientID", "env", env)
		return value, err
	}
	return value, nil
}

func GetPatClientID() (string, error) {
	envSecret := os.Getenv("SAIL_CLIENT_ID")
	if envSecret != "" {
		return envSecret, nil
	} else {
		value, err := GetClientID(GetActiveEnvironment())
		if err != nil {
			return value, err
		}
		return value, nil
	}
}

func GetClientSecret(env string) (string, error) {
	value, err := keyring.Get("environments.pat.clientsecret", env)
	if err != nil {
		log.Error("issue retrieving clientSecret", "env", env)
		return value, err
	}
	return value, nil
}

func GetPatClientSecret() (string, error) {
	envSecret := os.Getenv("SAIL_CLIENT_SECRET")
	if envSecret != "" {
		return envSecret, nil
	} else {
		value, err := GetClientSecret(GetActiveEnvironment())
		if err != nil {
			return value, err
		}
		return value, nil
	}
}

func SetPatClientID(ClientID string) error {
	err := keyring.Set("environments.pat.clientid", GetActiveEnvironment(), ClientID)
	if err != nil {
		return err
	}
	return nil
}

func SetPatClientSecret(ClientSecret string) error {
	err := keyring.Set("environments.pat.clientsecret", GetActiveEnvironment(), ClientSecret)
	if err != nil {
		return err
	}
	return nil
}

func PATLogin() (PATSet, error) {
	var set PATSet

	uri, err := url.Parse(GetTokenUrl())
	if err != nil {
		return set, err
	}

	query := &url.Values{}
	query.Add("grant_type", "client_credentials")
	uri.RawQuery = query.Encode()

	data := &url.Values{}

	patClientID, err := GetPatClientID()
	if err != nil {
		return set, err
	}
	patClientSecret, err := GetPatClientSecret()
	if err != nil {
		return set, err
	}

	data.Add("client_id", patClientID)
	data.Add("client_secret", patClientSecret)

	ctx := context.TODO()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return set, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return set, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return set, fmt.Errorf("failed to retrieve access token. status %s", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return set, err
	}

	var tResponse TokenResponse

	err = json.Unmarshal(raw, &tResponse)
	if err != nil {
		return set, err
	}

	now := time.Now()

	set = PATSet{AccessToken: tResponse.AccessToken, AccessExpiry: now.Add(time.Second * time.Duration(tResponse.ExpiresIn))}

	return set, nil
}

func PromptForClientID() (string, error) {
	const maxAttempts = 3
	var ClientID string
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Prompt for the Client ID
		ClientID, err = terminal.PromptPassword("Personal Access Token Client ID:")
		if err != nil {
			return "", err
		}

		log.Warn("ClientID", "length", len(ClientID))
		// Check length
		if len(ClientID) == 36 || len(ClientID) == 32 {
			fmt.Println("Valid Client ID provided.")
			return ClientID, nil
		} else {
			fmt.Printf("Invalid Client ID.Please ensure it is either 32 or 36 characters long Attempt %d/%d.\n", attempt, maxAttempts)
		}
	}

	return "", fmt.Errorf("maximum attempts reached for entering Client ID")
}

func PromptForClientSecret() (string, error) {
	const maxAttempts = 3
	var ClientSecret string
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Prompt for the Client Secret
		ClientSecret, err = terminal.PromptPassword("Personal Access Token Client Secret:")
		if err != nil {
			return "", err
		}

		// Check length
		if len(ClientSecret) == 64 {
			fmt.Println("Valid Client Secret provided.")
			return ClientSecret, nil
		} else {
			fmt.Printf("Invalid Client Secret. Please ensure it is 64 characters long. Attempt %d/%d.\n", attempt, maxAttempts)
		}
	}

	return "", fmt.Errorf("maximum attempts reached for entering Client Secret")
}
