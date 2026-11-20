package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	sailpoint "github.com/sailpoint-oss/golang-sdk/v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/viper"
	keyring "github.com/zalando/go-keyring"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Value sources reported by whoami.
const (
	SourceEnv     = "environment variable"
	SourceFlag    = "flag"
	SourceConfig  = "config file"
	SourceKeyring = "keychain"
	SourceDefault = "default"
	SourceNone    = "not set"
)

// ErrNotLoggedIn means there is no usable token and getting one would need an interactive login.
var ErrNotLoggedIn = fmt.Errorf("no cached OAuth session, and no refresh token that is still valid")

// TokenClaims are the Identity Security access token claims that whoami reports.
type TokenClaims struct {
	UserName   string   `json:"userName,omitempty"`
	IdentityID string   `json:"identityId,omitempty"`
	ClientID   string   `json:"clientId,omitempty"`
	Org        string   `json:"org,omitempty"`
	Pod        string   `json:"pod,omitempty"`
	Scopes     []string `json:"scopes,omitempty"`
	// ScopesEncoded is true when the token carries scopes in SailPoint's compact
	// encoded form (for example "Bg==") rather than as readable names.
	ScopesEncoded bool      `json:"scopesEncoded"`
	Authorities   []string  `json:"authorities,omitempty"`
	Internal      bool      `json:"internal"`
	StrongAuth    bool      `json:"strongAuth"`
	IssuedAt      time.Time `json:"issuedAt,omitempty"`
	ExpiresAt     time.Time `json:"expiresAt,omitempty"`
}

// HasUserContext reports whether the token is tied to an identity.
// Tokens from client credentials without a user (API clients) have no user_name.
func (c TokenClaims) HasUserContext() bool {
	return c.UserName != "" || c.IdentityID != ""
}

// AuthInfo describes how the CLI is authenticated and where each setting came from.
type AuthInfo struct {
	Environment       string `json:"environment"`
	EnvironmentSource string `json:"environmentSource"`
	AuthType          string `json:"authType"`
	AuthTypeSource    string `json:"authTypeSource"`
	BaseURL           string `json:"baseUrl"`
	BaseURLSource     string `json:"baseUrlSource"`
	TenantURL         string `json:"tenantUrl,omitempty"`
	TenantURLSource   string `json:"tenantUrlSource,omitempty"`

	// PAT only: where the client ID and secret were read from.
	ClientIDSource     string `json:"clientIdSource,omitempty"`
	ClientSecretSource string `json:"clientSecretSource,omitempty"`

	// Every SAIL_* environment variable that is set, with sensitive values masked.
	// These override the config file and keychain, so they are listed explicitly.
	EnvVars []EnvVar `json:"envVars"`

	LoggedIn   bool         `json:"loggedIn"`
	TokenError string       `json:"tokenError,omitempty"`
	Claims     *TokenClaims `json:"claims,omitempty"`

	// PAT only: the token record from the tenant, when it can be looked up.
	PAT            *PATDetails `json:"pat,omitempty"`
	PATLookupError string      `json:"patLookupError,omitempty"`

	// Things the user should probably act on, such as stale or overridden credentials.
	Warnings []string `json:"warnings"`
}

// EnvVar is a SAIL_* environment variable with its value masked if it is sensitive.
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PATDetails is the tenant's record of the personal access token the CLI is using.
type PATDetails struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Owner     string     `json:"owner,omitempty"`
	Created   time.Time  `json:"created"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Scopes    []string   `json:"scopes,omitempty"`
}

// envSet reports whether an environment variable is set to a non-empty value.
func envSet(name string) bool {
	v, ok := os.LookupEnv(name)
	return ok && strings.TrimSpace(v) != ""
}

// viperSource works out where viper got a key from. viper is configured with
// the "sail" prefix and AutomaticEnv, so key "authtype" maps to SAIL_AUTHTYPE.
func viperSource(key string) string {
	if envSet("SAIL_" + strings.ToUpper(key)) {
		return SourceEnv + " (SAIL_" + strings.ToUpper(key) + ")"
	}
	if viper.InConfig(key) {
		return SourceConfig
	}
	return SourceDefault
}

// secretSource reports whether a PAT credential comes from an env var or the keychain.
func secretSource(envName, keyringService string) string {
	if envSet(envName) {
		return SourceEnv + " (" + envName + ")"
	}
	if v, err := keyring.Get(keyringService, GetActiveEnvironment()); err == nil && v != "" {
		return SourceKeyring
	}
	return SourceNone
}

// showFullIDs turns off client ID masking. GetAuthInfo sets it.
var showFullIDs bool

// GetAuthInfo collects the current auth configuration without logging in interactively.
// Client IDs are masked unless debug was turned on for this run with --debug
// (debugFlagSet) or SAIL_DEBUG. A debug value saved in the config file doesn't
// count, because the CLI saves --debug to the config file after every command.
func GetAuthInfo(envFlagSet, debugFlagSet bool) (AuthInfo, error) {
	var info AuthInfo

	if err := InitConfig(); err != nil {
		return info, err
	}
	showFullIDs = GetDebug() && (debugFlagSet || envSet("SAIL_DEBUG"))

	info.Environment = GetActiveEnvironment()
	if envFlagSet {
		info.EnvironmentSource = SourceFlag + " (--env)"
	} else {
		info.EnvironmentSource = viperSource("activeenvironment")
	}

	info.AuthType = GetAuthType()
	info.AuthTypeSource = viperSource("authtype")

	info.BaseURL = GetBaseUrl()
	switch {
	case envSet("SAIL_BASE_URL"):
		info.BaseURLSource = SourceEnv + " (SAIL_BASE_URL)"
	case info.BaseURL != "":
		info.BaseURLSource = SourceConfig
	default:
		info.BaseURLSource = SourceNone
	}

	if info.AuthType == "oauth" {
		info.TenantURL = GetTenantUrl()
		if info.TenantURL != "" {
			info.TenantURLSource = SourceConfig
		} else {
			info.TenantURLSource = SourceNone
		}
	}

	info.EnvVars = collectEnvVars(os.Environ())
	info.Warnings = []string{}

	if info.AuthType == "pat" {
		info.ClientIDSource = secretSource("SAIL_CLIENT_ID", "environments.pat.clientid")
		info.ClientSecretSource = secretSource("SAIL_CLIENT_SECRET", "environments.pat.clientsecret")
	}

	info.Warnings = append(info.Warnings, overrideWarnings(info)...)

	token, tokenWarnings, err := GetAuthTokenNonInteractive()
	info.Warnings = append(info.Warnings, tokenWarnings...)
	if err != nil {
		info.TokenError = errorReason(err)
		if info.AuthType == "pat" && (envSet("SAIL_CLIENT_ID") || envSet("SAIL_CLIENT_SECRET")) {
			info.Warnings = append(info.Warnings, "Login failed with PAT credentials from environment variables. They may be out of date, revoked or for another tenant. Check or unset SAIL_CLIENT_ID and SAIL_CLIENT_SECRET.")
		}
		return info, nil
	}

	claims, err := ParseTokenClaims(token)
	if err != nil {
		info.TokenError = fmt.Sprintf("failed to decode access token: %v", err)
		return info, nil
	}

	// Mask the client ID like everywhere else in the output, since people paste
	// whoami output into tickets. ParseTokenClaims itself returns it unmasked.
	if claims.ClientID != "" {
		claims.ClientID = maskID(claims.ClientID)
	}

	info.LoggedIn = true
	info.Claims = &claims

	if info.AuthType == "pat" && claims.HasUserContext() {
		clientID, _ := GetPatClientID()
		pat, err := lookupPAT(token, clientID)
		if err != nil {
			info.PATLookupError = errorReason(err)
		} else {
			info.PAT = pat
		}
	}

	return info, nil
}

// htmlStartPattern finds the start of an HTML page in an error message.
var htmlStartPattern = regexp.MustCompile(`(?i)<!doctype|<html`)

// htmlTagPattern matches HTML tags, along with the contents of script and style blocks.
var htmlTagPattern = regexp.MustCompile(`(?is)<(script|style)\b.*?</(script|style)>|<[^>]+>`)

// errorReason returns an error's message without any HTML the server sent with it.
// When a response isn't JSON, such as a 404 page, the error has the whole body.
func errorReason(err error) string {
	msg := err.Error()
	if loc := htmlStartPattern.FindStringIndex(msg); loc != nil {
		if reason := strings.TrimRight(strings.TrimSpace(msg[:loc[0]]), ":"); reason != "" {
			return reason
		}
	}
	if !htmlTagPattern.MatchString(msg) {
		return strings.TrimSpace(msg)
	}
	return strings.Join(strings.Fields(htmlTagPattern.ReplaceAllString(msg, " ")), " ")
}

// sensitiveEnvVar reports whether a variable's value must never be printed.
func sensitiveEnvVar(name string) bool {
	upper := strings.ToUpper(name)
	return strings.Contains(upper, "SECRET") || strings.Contains(upper, "TOKEN") || strings.Contains(upper, "PASSWORD")
}

// maskID shows only the last four characters of an identifier, or all of it
// when debug was turned on for this run.
func maskID(v string) string {
	if showFullIDs {
		return v
	}
	if len(v) <= 4 {
		return strings.Repeat("*", len(v))
	}
	return "****" + v[len(v)-4:]
}

// collectEnvVars returns the SAIL_* variables in environ, sorted, with sensitive values masked.
func collectEnvVars(environ []string) []EnvVar {
	vars := []EnvVar{}
	for _, kv := range environ {
		name, value, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(strings.ToUpper(name), "SAIL_") || strings.TrimSpace(value) == "" {
			continue
		}
		switch {
		case sensitiveEnvVar(name):
			value = "(set, hidden)"
		case strings.HasSuffix(strings.ToUpper(name), "CLIENT_ID"):
			value = maskID(value)
		}
		vars = append(vars, EnvVar{Name: name, Value: value})
	}
	sort.Slice(vars, func(i, j int) bool { return vars[i].Name < vars[j].Name })
	return vars
}

// overrideWarnings flags env vars that replace a different value saved for the active environment.
// A stale env var silently winning over saved config is easy to miss, so whoami calls it out.
func overrideWarnings(info AuthInfo) []string {
	env := GetActiveEnvironment()
	var stored []override

	if v := os.Getenv("SAIL_BASE_URL"); v != "" {
		stored = append(stored, override{"SAIL_BASE_URL", "base URL", v, GetEnvBaseUrl(env)})
	}
	if info.AuthType == "pat" {
		if v := os.Getenv("SAIL_CLIENT_ID"); v != "" {
			saved, _ := keyring.Get("environments.pat.clientid", env)
			stored = append(stored, override{"SAIL_CLIENT_ID", "PAT client ID", v, saved})
		}
		if v := os.Getenv("SAIL_CLIENT_SECRET"); v != "" {
			saved, _ := keyring.Get("environments.pat.clientsecret", env)
			stored = append(stored, override{"SAIL_CLIENT_SECRET", "PAT client secret", v, saved})
		}
	}

	return describeOverrides(env, stored)
}

// override is an env var value and the value saved for the same setting.
type override struct {
	envName, label, envValue, savedValue string
}

// describeOverrides returns a warning for each env var that differs from a saved
// value. Warnings never include either value.
func describeOverrides(env string, overrides []override) []string {
	var warnings []string
	for _, o := range overrides {
		if o.savedValue == "" || o.savedValue == o.envValue {
			continue
		}
		warnings = append(warnings, fmt.Sprintf("%s is overriding the %s saved for environment %q. If that's not intended, unset %s.", o.envName, o.label, env, o.envName))
	}
	return warnings
}

// lookupPAT finds the caller's personal access token whose ID matches the client ID in use.
func lookupPAT(token, clientID string) (*PATDetails, error) {
	if clientID == "" {
		return nil, fmt.Errorf("no PAT client ID available")
	}

	configuration := sailpoint.NewCLIConfiguration(sailpoint.ClientConfiguration{Token: token, BaseURL: GetBaseUrl()})
	// NewAPIClient creates configuration.HTTPClient, so the logger must be set after it.
	apiClient := sailpoint.NewAPIClient(configuration)
	var devNull types.DevNull
	configuration.HTTPClient.Logger = devNull

	pats, _, err := apiClient.PersonalAccessTokensAPI.ListPersonalAccessTokensV1(context.TODO()).OwnerId("me").Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list personal access tokens: %w", err)
	}

	for _, p := range pats {
		if !strings.EqualFold(p.Id, clientID) {
			continue
		}
		details := &PATDetails{
			ID:      maskID(p.Id),
			Name:    p.Name,
			Created: p.Created.Time,
			Scopes:  p.Scope,
		}
		if p.Owner.Name != nil {
			details.Owner = *p.Owner.Name
		}
		if exp := p.ExpirationDate.Get(); exp != nil && !exp.Time.IsZero() {
			t := exp.Time
			details.ExpiresAt = &t
		}
		return details, nil
	}

	return nil, fmt.Errorf("the PAT in use was not found among your personal access tokens")
}

// GetAuthTokenNonInteractive returns an access token for the active environment
// and any warnings about the token cache. It works like GetAuthToken, except it
// never starts a browser login.
//
// For PAT, the cached token is used only when it was issued to the configured
// client ID and the credentials don't come from environment variables. Otherwise
// it requests a new token, and caches it only when the credentials are the saved
// ones. For OAuth, it refreshes the token when the refresh token is still valid,
// and returns ErrNotLoggedIn when a login would be needed.
func GetAuthTokenNonInteractive() (string, []string, error) {
	if err := Validate(); err != nil {
		return "", nil, err
	}

	now := time.Now()

	switch GetAuthType() {
	case "pat":
		var warnings []string
		clientID, _ := GetPatClientID()
		// Credentials from env vars are always tested with a fresh login, because a
		// cached token says nothing about whether the env var secret works.
		envCreds := envSet("SAIL_CLIENT_ID") || envSet("SAIL_CLIENT_SECRET")

		if expiry, err := GetPatTokenExpiry(); err == nil && expiry.After(now) {
			if token, err := GetPatToken(); err == nil && token != "" {
				matches, cachedID := cachedTokenMatches(token, clientID)
				if matches && !envCreds {
					return token, nil, nil
				}
				if !matches {
					warnings = append(warnings, fmt.Sprintf("The keychain has a cached token for environment %q that was issued to a different PAT (%s). Other sail commands will keep using that token instead of the configured credentials until it expires at %s.", GetActiveEnvironment(), maskID(cachedID), expiry.Local().Format(time.RFC3339)))
				}
			}
		}

		set, err := PATLogin()
		if err != nil {
			return "", warnings, err
		}
		if envCreds {
			// Don't let whoami replace the saved session with an env var login.
			return set.AccessToken, warnings, nil
		}
		CachePAT(set)
		return set.AccessToken, warnings, nil

	case "oauth":
		if expiry, err := GetOAuthTokenExpiry(); err == nil && expiry.After(now) {
			if token, err := GetOAuthToken(); err == nil && token != "" {
				return token, nil, nil
			}
		}
		if refreshExpiry, err := GetOAuthRefreshExpiry(); err == nil && refreshExpiry.After(now) {
			set, err := RefreshOAuth()
			if err != nil {
				return "", nil, err
			}
			CacheOAuth(set)
			return set.AccessToken, nil, nil
		}
		return "", nil, ErrNotLoggedIn

	default:
		return "", nil, fmt.Errorf("invalid authtype configured")
	}
}

// cachedTokenMatches reports whether a cached PAT token was issued to clientID,
// and returns the client ID the token was actually issued to. A token that
// can't be decoded, or has no client ID, is treated as a match so behavior
// stays the same as GetAuthToken.
func cachedTokenMatches(token, clientID string) (bool, string) {
	claims, err := ParseTokenClaims(token)
	if err != nil || claims.ClientID == "" || clientID == "" {
		return true, claims.ClientID
	}
	return strings.EqualFold(claims.ClientID, clientID), claims.ClientID
}

// rawClaims mirrors the JSON claims in an Identity Security access token.
type rawClaims struct {
	UserName    string           `json:"user_name"`
	IdentityID  string           `json:"identity_id"`
	ClientID    string           `json:"client_id"`
	Org         string           `json:"org"`
	Pod         string           `json:"pod"`
	Scope       interface{}      `json:"scope"`
	Authorities []string         `json:"authorities"`
	Internal    bool             `json:"internal"`
	StrongAuth  bool             `json:"strong_auth"`
	IssuedAt    *jwt.NumericDate `json:"iat"`
	Expiry      *jwt.NumericDate `json:"exp"`
}

// ParseTokenClaims decodes an access token's claims without verifying its signature.
// Only use it to display information; never use it to make authorization decisions.
func ParseTokenClaims(tokenString string) (TokenClaims, error) {
	var claims TokenClaims

	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return claims, err
	}

	var raw rawClaims
	if err := token.UnsafeClaimsWithoutVerification(&raw); err != nil {
		return claims, err
	}

	claims = TokenClaims{
		UserName:    raw.UserName,
		IdentityID:  raw.IdentityID,
		ClientID:    raw.ClientID,
		Org:         raw.Org,
		Pod:         raw.Pod,
		Scopes:      normalizeScopes(raw.Scope),
		Authorities: raw.Authorities,
		Internal:    raw.Internal,
		StrongAuth:  raw.StrongAuth,
	}
	claims.ScopesEncoded = scopesEncoded(claims.Scopes)
	if raw.IssuedAt != nil {
		claims.IssuedAt = raw.IssuedAt.Time()
	}
	if raw.Expiry != nil {
		claims.ExpiresAt = raw.Expiry.Time()
	}

	return claims, nil
}

// normalizeScopes accepts the scope claim as either a space-delimited string or a list.
func normalizeScopes(scope interface{}) []string {
	switch v := scope.(type) {
	case string:
		return strings.Fields(v)
	case []interface{}:
		scopes := make([]string, 0, len(v))
		for _, s := range v {
			if str, ok := s.(string); ok && str != "" {
				scopes = append(scopes, str)
			}
		}
		return scopes
	default:
		return nil
	}
}

// scopesEncoded reports whether every scope looks like SailPoint's compact encoding
// rather than a readable name. Readable scopes contain a colon (sp:scopes:all);
// encoded ones are base64 strings such as "Bg==".
func scopesEncoded(scopes []string) bool {
	if len(scopes) == 0 {
		return false
	}
	for _, s := range scopes {
		if strings.Contains(s, ":") {
			return false
		}
		if _, err := base64.StdEncoding.DecodeString(s); err != nil {
			return false
		}
	}
	return true
}
