package auth

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/keyring"
	"gopkg.in/square/go-jose.v2/jwt"
)

// GetToken obtains a valid access token for the given environment.
// It checks cached tokens first, refreshes if possible, and performs a full
// login only when necessary. The setBaseURL callback is invoked if the OAuth
// flow provides an updated base URL.
func GetToken(authType, env, baseURL, tenantURL, tokenURL string, setBaseURL func(string)) (string, error) {
	if err := ValidateAuth(authType, env, baseURL, tenantURL); err != nil {
		return "", err
	}

	var token string

	switch authType {
	case "pat":
		t, err := getPatToken(env, tokenURL)
		if err != nil {
			return "", err
		}
		token = t

	case "oauth":
		t, err := getOAuthToken(env, baseURL, tenantURL, setBaseURL)
		if err != nil {
			return "", err
		}
		token = t

	default:
		return "", fmt.Errorf("invalid authtype '%s' configured", authType)
	}

	if err := CheckToken(token); err != nil {
		return "", err
	}

	return token, nil
}

func getPatToken(env, tokenURL string) (string, error) {
	authExpiry, _ := GetPatTokenExpiry(env)

	if authExpiry.After(time.Now()) {
		return GetPatToken(env)
	}

	clientID, err := GetPatClientID(env)
	if err != nil {
		return "", err
	}
	clientSecret, err := GetPatClientSecret(env)
	if err != nil {
		return "", err
	}

	set, err := PATLogin(tokenURL, clientID, clientSecret)
	if err != nil {
		return "", err
	}

	if err := CachePAT(env, set); err != nil {
		log.Warn("Failed to cache PAT token in keyring", "error", err)
	}

	return set.AccessToken, nil
}

func getOAuthToken(env, baseURL, tenantURL string, setBaseURL func(string)) (string, error) {
	authExpiry, _ := GetOAuthTokenExpiry(env)
	refreshExpiry, _ := GetOAuthRefreshExpiry(env)

	if authExpiry.After(time.Now()) {
		return GetOAuthToken(env)
	}

	if refreshExpiry.After(time.Now()) {
		set, err := RefreshOAuth(env, baseURL, tenantURL)
		if err != nil {
			return "", err
		}
		if err := CacheOAuth(env, set); err != nil {
			log.Warn("Failed to cache OAuth token in keyring", "error", err)
		}
		return set.AccessToken, nil
	}

	set, err := OAuthLogin(baseURL)
	if err != nil {
		return "", err
	}

	if set.BaseURL != "" && setBaseURL != nil {
		setBaseURL(set.BaseURL)
	}

	if err := CacheOAuth(env, set); err != nil {
		log.Warn("Failed to cache OAuth token in keyring", "error", err)
	}

	return set.AccessToken, nil
}

// ValidateAuth checks that the necessary configuration exists for the given auth type.
func ValidateAuth(authType, env, baseURL, tenantURL string) error {
	var errors int

	supportsSecrets := keyring.IsSupported()

	switch authType {
	case "pat":
		if !supportsSecrets {
			log.Warn("Secrets storage is not currently functional on this platform, PAT will only work with environment variables")
		}
		if baseURL == "" {
			log.Error("configured environment is missing BaseURL")
			errors++
		}
		patClientID, err := GetPatClientID(env)
		if err != nil {
			return err
		}
		patClientSecret, err := GetPatClientSecret(env)
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
		if !supportsSecrets {
			log.Warn("Secrets storage is not currently functional on this platform, every command will reauthenticate with OAuth")
		}
		if baseURL == "" {
			log.Error("configured environment is missing BaseURL")
			errors++
		}
		if tenantURL == "" {
			log.Error("configured environment is missing TenantURL")
			errors++
		}

	default:
		log.Error("invalid authtype configured", "authtype", authType)
		errors++
	}

	if errors > 0 {
		return fmt.Errorf("configuration invalid, errors: %v", errors)
	}

	return nil
}

// CheckToken parses a JWT and logs identity info, warning if user context is missing.
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

// GetTokenClaims parses a JWT and returns the claims map (for status display).
func GetTokenClaims(tokenString string) (map[string]interface{}, error) {
	var claims map[string]interface{}

	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return nil, err
	}

	token.UnsafeClaimsWithoutVerification(&claims)
	return claims, nil
}

// Logout clears all cached tokens for the given environment and auth type.
func Logout(authType, env string) error {
	switch authType {
	case "pat":
		return ResetCachePAT(env)
	case "oauth":
		return ResetCacheOAuth(env)
	default:
		return fmt.Errorf("unknown auth type: %s", authType)
	}
}
