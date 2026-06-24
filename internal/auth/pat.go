package auth

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
	"github.com/sailpoint-oss/sailpoint-cli/internal/keyring"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
)

const (
	patClientIDService     = "environments.pat.clientid"
	patClientSecretService = "environments.pat.clientsecret"
	patAccessTokenService  = "environments.pat.accesstoken"
	patExpiryService       = "environments.pat.expiry"
)

// GetPatClientID returns the PAT client ID for the given environment,
// checking the SAIL_CLIENT_ID env var first.
func GetPatClientID(env string) (string, error) {
	if v := os.Getenv("SAIL_CLIENT_ID"); v != "" {
		return v, nil
	}
	return keyring.Get(patClientIDService, env)
}

// GetPatClientSecret returns the PAT client secret for the given environment,
// checking the SAIL_CLIENT_SECRET env var first.
func GetPatClientSecret(env string) (string, error) {
	if v := os.Getenv("SAIL_CLIENT_SECRET"); v != "" {
		return v, nil
	}
	return keyring.Get(patClientSecretService, env)
}

// SetPatClientID stores the PAT client ID in the keyring.
func SetPatClientID(env, clientID string) error {
	return keyring.Set(patClientIDService, env, clientID)
}

// SetPatClientSecret stores the PAT client secret in the keyring.
func SetPatClientSecret(env, clientSecret string) error {
	return keyring.Set(patClientSecretService, env, clientSecret)
}

// DeletePatClientID removes the PAT client ID from the keyring.
func DeletePatClientID(env string) error {
	return keyring.Delete(patClientIDService, env)
}

// DeletePatClientSecret removes the PAT client secret from the keyring.
func DeletePatClientSecret(env string) error {
	return keyring.Delete(patClientSecretService, env)
}

// GetPatToken retrieves the cached PAT access token from the keyring.
func GetPatToken(env string) (string, error) {
	return keyring.Get(patAccessTokenService, env)
}

// GetPatTokenExpiry retrieves the cached PAT token expiry from the keyring.
func GetPatTokenExpiry(env string) (time.Time, error) {
	return keyring.GetTime(patExpiryService, env)
}

// CachePAT stores the PAT access token and expiry in the keyring.
func CachePAT(env string, set PATSet) error {
	if err := keyring.Set(patAccessTokenService, env, set.AccessToken); err != nil {
		return err
	}
	return keyring.SetTime(patExpiryService, env, set.AccessExpiry)
}

// ResetCachePAT clears the cached PAT token and expiry from the keyring.
func ResetCachePAT(env string) error {
	token, err := GetPatToken(env)
	if token != "" && err == nil {
		if err := keyring.Delete(patAccessTokenService, env); err != nil {
			return err
		}
	}
	expiry, err := GetPatTokenExpiry(env)
	if !expiry.IsZero() && err == nil {
		if err := keyring.Delete(patExpiryService, env); err != nil {
			return err
		}
	}
	return nil
}

// DeleteAllPatSecrets removes all PAT-related keyring entries for an environment.
func DeleteAllPatSecrets(env string) {
	_ = keyring.Delete(patAccessTokenService, env)
	_ = keyring.Delete(patExpiryService, env)
	_ = keyring.Delete(patClientIDService, env)
	_ = keyring.Delete(patClientSecretService, env)
}

// PATLogin performs a client_credentials token exchange and returns the token set.
func PATLogin(tokenURL, clientID, clientSecret string) (PATSet, error) {
	var set PATSet

	uri, err := url.Parse(tokenURL)
	if err != nil {
		return set, err
	}

	query := &url.Values{}
	query.Add("grant_type", "client_credentials")
	uri.RawQuery = query.Encode()

	data := &url.Values{}
	data.Add("client_id", clientID)
	data.Add("client_secret", clientSecret)

	ctx := context.TODO()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return set, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := (&http.Client{}).Do(req)
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
	if err := json.Unmarshal(raw, &tResponse); err != nil {
		return set, err
	}

	set = PATSet{
		AccessToken:  tResponse.AccessToken,
		AccessExpiry: time.Now().Add(time.Second * time.Duration(tResponse.ExpiresIn)),
	}
	return set, nil
}

// PromptForClientID interactively prompts for a PAT Client ID with validation.
func PromptForClientID() (string, error) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		clientID, err := tui.Password("Personal Access Token Client ID")
		if err != nil {
			return "", err
		}
		if len(clientID) == 36 || len(clientID) == 32 {
			log.Debug("Valid Client ID entered", "length", len(clientID))
			return clientID, nil
		}
		log.Warn("Invalid Client ID length", "got", len(clientID), "expected", "32 or 36", "attempt", fmt.Sprintf("%d/%d", attempt, maxAttempts))
	}
	return "", fmt.Errorf("maximum attempts reached for entering Client ID")
}

// PromptForClientSecret interactively prompts for a PAT Client Secret with validation.
func PromptForClientSecret() (string, error) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		clientSecret, err := tui.Password("Personal Access Token Client Secret")
		if err != nil {
			return "", err
		}
		if len(clientSecret) == 64 {
			log.Debug("Valid Client Secret entered")
			return clientSecret, nil
		}
		log.Warn("Invalid Client Secret length", "got", len(clientSecret), "expected", 64, "attempt", fmt.Sprintf("%d/%d", attempt, maxAttempts))
	}
	return "", fmt.Errorf("maximum attempts reached for entering Client Secret")
}
