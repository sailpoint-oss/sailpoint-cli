package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	keyring "github.com/zalando/go-keyring"
	"gopkg.in/square/go-jose.v2/jwt"
)

// accessTokenEnvVar carries an Identity Security Cloud access token that was
// already minted at the OAuth token endpoint. When it is set, the CLI sends
// that token on every request and does not exchange a client ID and secret.
const accessTokenEnvVar = "SAIL_ACCESS_TOKEN"

const bearerTokenService = "environments.bearer.accesstoken"

// runtimeAccessToken is the --access-token value for this process. It is kept
// out of viper so SaveConfig does not write the token into the config file.
var runtimeAccessToken string

// SetRuntimeAccessToken records the access token passed with --access-token.
func SetRuntimeAccessToken(token string) {
	runtimeAccessToken = strings.TrimSpace(token)
}

// RuntimeAccessTokenVar is the destination for the --access-token flag.
// The flag package writes it when the command line is parsed, so the token
// stays in memory and is not saved with the config file.
func RuntimeAccessTokenVar() *string {
	return &runtimeAccessToken
}

// authTypeInUse reports the authentication method for this process.
// An explicitly supplied access token selects bearer authentication so a
// deployment can pass the minted token without also storing a client ID and
// secret.
func authTypeInUse() string {
	if _, ok := explicitBearerToken(); ok {
		return "bearer"
	}
	return GetAuthType()
}

// explicitBearerToken returns a token supplied for this process, rather than
// one stored in the keyring.
func explicitBearerToken() (string, bool) {
	if env := strings.TrimSpace(os.Getenv(accessTokenEnvVar)); env != "" {
		return env, true
	}
	if runtimeAccessToken != "" {
		return runtimeAccessToken, true
	}
	return "", false
}

// GetBearerToken returns the access token to send as a bearer token.
// SAIL_ACCESS_TOKEN and --access-token take precedence over a token stored
// for the active environment. The token is used as-is: this does not call
// the OAuth token endpoint.
func GetBearerToken() (string, error) {
	if raw, ok := explicitBearerToken(); ok {
		token, err := parseBearerToken(raw)
		if err != nil {
			return "", err
		}
		return token, nil
	}

	env := GetActiveEnvironment()
	value, err := keyring.Get(bearerTokenService, env)
	if err != nil {
		return "", missingBearerTokenError(env, err)
	}

	token, err := parseBearerToken(value)
	if err != nil {
		return "", err
	}
	return token, nil
}

// SetBearerToken stores an access token for the active environment.
func SetBearerToken(token string) error {
	parsed, err := parseBearerToken(token)
	if err != nil {
		return err
	}

	err = keyring.Set(bearerTokenService, GetActiveEnvironment(), parsed)
	if err != nil {
		return err
	}
	return nil
}

// PromptForBearerToken prompts for an access token that was already minted at
// the Identity Security Cloud OAuth token endpoint.
func PromptForBearerToken() (string, error) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		token, err := terminal.PromptMasked("Access token:")
		if err != nil {
			return "", err
		}

		parsed, err := parseBearerToken(token)
		if err == nil {
			return parsed, nil
		}

		fmtBearerAttempt(err, attempt, maxAttempts)
	}

	return "", fmt.Errorf("maximum attempts reached for entering the bearer token")
}

// parseBearerToken accepts either the raw access token or a value prefixed
// with "Bearer ". It rejects tokens that are not signed JWTs or that have
// already expired. An expired token cannot be refreshed here, because the
// client ID and secret that minted it are not available.
func parseBearerToken(raw string) (string, error) {
	token := normalizeBearerToken(raw)
	if token == "" {
		return "", fmt.Errorf("the bearer token is empty")
	}

	parsed, err := jwt.ParseSigned(token)
	if err != nil {
		return "", fmt.Errorf("the bearer token is not a signed Identity Security Cloud access token: %w\n\n"+
			"Use the access_token returned by the OAuth token endpoint. The CLI sends that token on each request and does not exchange a client ID and secret", err)
	}

	var claims jwt.Claims
	if err := parsed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return "", fmt.Errorf("the bearer token is not a signed Identity Security Cloud access token: %w", err)
	}

	if claims.Expiry != nil && !claims.Expiry.Time().After(time.Now()) {
		return "", expiredBearerTokenError(claims.Expiry.Time())
	}

	return token, nil
}

func normalizeBearerToken(raw string) string {
	token := strings.TrimSpace(raw)
	if len(token) >= 2 {
		if (token[0] == '"' && token[len(token)-1] == '"') || (token[0] == '\'' && token[len(token)-1] == '\'') {
			token = strings.TrimSpace(token[1 : len(token)-1])
		}
	}
	if len(token) >= 7 && strings.EqualFold(token[:7], "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	return token
}

func expiredBearerTokenError(expiry time.Time) error {
	return fmt.Errorf("the bearer token expired at %s\n\n"+
		"Mint a new access token at the Identity Security Cloud OAuth token endpoint, then set %s, pass --access-token, or run `sail set bearer`",
		expiry.UTC().Format(time.RFC3339), accessTokenEnvVar)
}

func missingBearerTokenError(env string, err error) error {
	guidance := fmt.Sprintf("Set %s, pass --access-token, or run `sail set bearer`", accessTokenEnvVar)
	if errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("no bearer token is stored for environment %q\n\n"+
			"Pass an access token that was already minted at the Identity Security Cloud OAuth token endpoint. %s", env, guidance)
	}
	return fmt.Errorf("failed to read the bearer token for environment %q from the keyring: %w\n\n%s", env, err, guidance)
}

func fmtBearerAttempt(err error, attempt, maxAttempts int) {
	message := err.Error()
	if i := strings.IndexByte(message, '\n'); i >= 0 {
		message = message[:i]
	}
	fmt.Printf("%s Attempt %d/%d.\n", message, attempt, maxAttempts)
}
