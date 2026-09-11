package config

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/skratchdot/open-golang/open"
	keyring "github.com/zalando/go-keyring"
	"golang.org/x/term"
	"gopkg.in/square/go-jose.v2/jwt"
)

type RefreshResponse struct {
	AccessToken         string `json:"access_token"`
	TokenType           string `json:"token_type"`
	RefreshToken        string `json:"refresh_token"`
	ExpiresIn           int    `json:"expires_in"`
	Scope               string `json:"scope"`
	TenantID            string `json:"tenant_id"`
	Internal            bool   `json:"internal"`
	Pod                 string `json:"pod"`
	StrongAuthSupported bool   `json:"strong_auth_supported"`
	Org                 string `json:"org"`
	ClaimsSupported     bool   `json:"claims_supported"`
	IdentityID          string `json:"identity_id"`
	StrongAuth          bool   `json:"strong_auth"`
	Jti                 string `json:"jti"`
}

type TokenSet struct {
	AccessToken   string
	AccessExpiry  time.Time
	RefreshToken  string
	RefreshExpiry time.Time
}

// oauthInfo is the tenant OAuth discovery document served at
// {baseURL}/oauth/info.
type oauthInfo struct {
	AuthorizeEndpoint string `json:"authorizeEndpoint"`
}

// pastePayload is the value the user copies from the redirect page. The page
// packs the authorization code together with the state it received, so this
// CLI can prove that the code belongs to the request it started.
type pastePayload struct {
	Version int    `json:"v"`
	Code    string `json:"code"`
	State   string `json:"state"`
}

func DeleteOAuthToken(env string) error {
	if env != "" {
		err := keyring.Delete("environments.oauth.accesstoken", env)
		if err != nil {
			return err
		}
		return nil
	} else {
		err := keyring.Delete("environments.oauth.accesstoken", GetActiveEnvironment())
		if err != nil {
			return err
		}
		return nil
	}
}

func GetOAuthToken() (string, error) {
	value, err := keyring.Get("environments.oauth.accesstoken", GetActiveEnvironment())
	if err != nil {
		return value, err
	}
	return value, nil
}

func SetOAuthToken(token string) error {
	err := keyring.Set("environments.oauth.accesstoken", GetActiveEnvironment(), token)
	if err != nil {
		return err
	}
	return nil
}

func DeleteOAuthTokenExpiry(env string) error {
	if env != "" {
		err := keyring.Delete("environments.oauth.expiry", env)
		if err != nil {
			return err
		}
		return nil
	} else {
		err := keyring.Delete("environments.oauth.expiry", GetActiveEnvironment())
		if err != nil {
			return err
		}
		return nil
	}
}

func GetOAuthTokenExpiry() (time.Time, error) {
	var valueTime time.Time
	valueString, err := keyring.Get("environments.oauth.expiry", GetActiveEnvironment())
	if err != nil {
		return valueTime, err
	}

	valueTime, err = GetTime(valueString)
	if err != nil {
		return valueTime, err
	}

	return valueTime, nil
}

func SetOAuthTokenExpiry(expiry time.Time) error {
	err := keyring.Set("environments.oauth.expiry", GetActiveEnvironment(), SetTime(expiry))
	if err != nil {
		return err
	}
	return nil
}

func DeleteRefreshToken(env string) error {
	if env != "" {
		err := keyring.Delete("environments.oauth.refreshtoken", env)
		if err != nil {
			return err
		}
		return nil
	} else {
		err := keyring.Delete("environments.oauth.refreshtoken", GetActiveEnvironment())
		if err != nil {
			return err
		}
		return nil
	}
}

func GetRefreshToken() (string, error) {
	value, err := keyring.Get("environments.oauth.refreshtoken", GetActiveEnvironment())

	if err != nil {
		return value, err
	}

	return value, nil
}

func SetRefreshToken(token string) error {

	err := keyring.Set("environments.oauth.refreshtoken", GetActiveEnvironment(), token)
	if err != nil {
		return err
	}

	return nil

}

func DeleteRefreshTokenExpiry(env string) error {
	if env != "" {
		err := keyring.Delete("environments.oauth.refreshexpiry", env)
		if err != nil {
			return err
		}
		return nil
	} else {
		err := keyring.Delete("environments.oauth.refreshexpiry", GetActiveEnvironment())
		if err != nil {
			return err
		}
		return nil
	}
}

func GetOAuthRefreshExpiry() (time.Time, error) {

	var valueTime time.Time
	valueString, err := keyring.Get("environments.oauth.refreshexpiry", GetActiveEnvironment())
	if err != nil {
		return valueTime, err
	}

	valueTime, err = GetTime(valueString)
	if err != nil {
		return valueTime, err
	}

	return valueTime, nil

}

func SetOAuthRefreshExpiry(expiry time.Time) error {

	err := keyring.Set("environments.oauth.refreshexpiry", GetActiveEnvironment(), SetTime(expiry))
	if err != nil {
		return err
	}

	return nil

}

const (
	// ClientID is the public OAuth client registered for the SailPoint
	// developer tools. Public clients hold no secret and must use PKCE.
	ClientID = "sailapps"

	// RedirectURI is a static page that displays the authorization code for
	// the user to copy.
	RedirectURI = "https://developer.sailpoint.com/sailapps"

	// redirectURLEnvVar overrides RedirectURI for local development. Only a
	// loopback URL is accepted, so the code can only reach this machine.
	redirectURLEnvVar = "SAIL_OAUTH_REDIRECT_URL"

	// pastePrefix marks a value produced by the redirect page.
	pastePrefix = "sp1."

	// pasteVersion is the payload version this CLI understands.
	pasteVersion = 1
)

// oauthHTTPClient is used for every OAuth request, so that a slow endpoint
// cannot block sign-in forever.
var oauthHTTPClient = &http.Client{Timeout: 30 * time.Second}

func ResetCacheOAuth() error {
	err := DeleteOAuthToken("")
	if err != nil {
		return err
	}

	err = DeleteOAuthTokenExpiry("")
	if err != nil {
		return err
	}

	err = DeleteRefreshToken("")
	if err != nil {
		return err
	}

	err = DeleteRefreshTokenExpiry("")
	if err != nil {
		return err
	}

	return nil
}

func CacheOAuth(set TokenSet) error {
	var err error

	err = SetOAuthToken(set.AccessToken)
	if err != nil {
		return err
	}

	err = SetOAuthTokenExpiry(set.AccessExpiry)
	if err != nil {
		return err
	}

	err = SetRefreshToken(set.RefreshToken)
	if err != nil {
		return err
	}

	err = SetOAuthRefreshExpiry(set.RefreshExpiry)
	if err != nil {
		return err
	}

	return nil
}

// redirectURI returns the redirect URI for this sign-in. The default is the
// static SailPoint page. SAIL_OAUTH_REDIRECT_URL overrides it for local
// development, and only a loopback host is accepted.
func redirectURI() (string, error) {
	override := strings.TrimSpace(os.Getenv(redirectURLEnvVar))
	if override == "" {
		return RedirectURI, nil
	}

	parsed, err := url.Parse(override)
	if err != nil {
		return "", fmt.Errorf("%s is not a valid URL: %v", redirectURLEnvVar, err)
	}

	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme != "http" || (host != "localhost" && host != "127.0.0.1") {
		return "", fmt.Errorf("%s must be a loopback URL, such as http://localhost:4200/sailapps", redirectURLEnvVar)
	}

	return override, nil
}

// assertHTTPSURL parses rawURL and rejects it unless it is a plain HTTPS URL.
// The host itself is not restricted, because a tenant can use a vanity domain.
func assertHTTPSURL(rawURL string, label string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid URL: %v", label, err)
	}
	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("%s must use HTTPS", label)
	}
	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("%s has no host", label)
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("%s must not include credentials", label)
	}
	if parsed.Fragment != "" {
		return nil, fmt.Errorf("%s must not include a fragment", label)
	}

	return parsed, nil
}

// discoverAuthorizeEndpoint reads the authorize endpoint from
// {baseURL}/oauth/info.
//
// The token endpoint is not taken from this document. The authorization code
// and the PKCE verifier always go to the tenant URL from configuration, so a
// discovery response can never move them to another host.
func discoverAuthorizeEndpoint(baseURL string) (string, error) {
	resp, err := oauthHTTPClient.Get(baseURL + "/oauth/info")
	if err != nil {
		return "", fmt.Errorf("failed to read tenant OAuth information: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tenant OAuth information returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var info oauthInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("failed to decode tenant OAuth information: %v", err)
	}

	if info.AuthorizeEndpoint == "" {
		return "", fmt.Errorf("tenant OAuth information is missing the authorize endpoint")
	}
	if _, err := assertHTTPSURL(info.AuthorizeEndpoint, "authorize endpoint"); err != nil {
		return "", err
	}

	return info.AuthorizeEndpoint, nil
}

// randomURLSafeString returns byteLength random bytes as a base64url string.
func randomURLSafeString(byteLength int) (string, error) {
	buf := make([]byte, byteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate random value: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// codeChallenge returns the S256 PKCE challenge for a verifier (RFC 7636).
func codeChallenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

// confirmationCodeFromState returns the short code that both this CLI and the
// redirect page display. The user compares the two before pasting.
func confirmationCodeFromState(state string) string {
	if len(state) < 8 {
		return ""
	}
	return state[:4] + "-" + state[4:8]
}

// parsePasteCode unpacks the value the user copied from the redirect page and
// verifies that its state matches the state this CLI sent.
func parsePasteCode(pasted string, expectedState string) (string, error) {
	pasted = strings.TrimSpace(pasted)
	if pasted == "" {
		return "", fmt.Errorf("no code was entered")
	}
	if !strings.HasPrefix(pasted, pastePrefix) {
		return "", fmt.Errorf("the code must start with %q, so it did not come from the SailPoint sign-in page", pastePrefix)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(pasted, pastePrefix))
	if err != nil {
		return "", fmt.Errorf("the code is damaged, so copy it again: %v", err)
	}

	var payload pastePayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return "", fmt.Errorf("the code is damaged, so copy it again: %v", err)
	}
	if payload.Version != pasteVersion {
		return "", fmt.Errorf("the code uses version %d, so update the CLI", payload.Version)
	}
	if payload.Code == "" {
		return "", fmt.Errorf("the code is missing the authorization code")
	}
	if subtle.ConstantTimeCompare([]byte(payload.State), []byte(expectedState)) != 1 {
		return "", fmt.Errorf("the code belongs to a different sign-in attempt, so start again")
	}

	return payload.Code, nil
}

// requestToken posts a form to the tenant token endpoint and returns the token
// response. The client authenticates with its client ID only, because the
// client is public.
func requestToken(tokenEndpoint string, form url.Values) (RefreshResponse, error) {
	var response RefreshResponse

	form.Set("client_id", ClientID)

	resp, err := oauthHTTPClient.PostForm(tokenEndpoint, form)
	if err != nil {
		return response, fmt.Errorf("failed to reach the token endpoint: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, fmt.Errorf("failed to read the token response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return response, fmt.Errorf("failed to decode the token response: %v", err)
	}
	if response.AccessToken == "" {
		return response, fmt.Errorf("no access token in the token response")
	}
	if response.RefreshToken == "" {
		return response, fmt.Errorf("no refresh token in the token response")
	}

	return response, nil
}

// tokenSetFromResponse reads the expiry of each token from its own claims.
func tokenSetFromResponse(response RefreshResponse) (TokenSet, error) {
	var set TokenSet

	accessExpiry, err := tokenExpiry(response.AccessToken)
	if err != nil {
		return set, fmt.Errorf("failed to parse access token: %v", err)
	}

	refreshExpiry, err := tokenExpiry(response.RefreshToken)
	if err != nil {
		return set, fmt.Errorf("failed to parse refresh token: %v", err)
	}

	return TokenSet{
		AccessToken:   response.AccessToken,
		AccessExpiry:  accessExpiry,
		RefreshToken:  response.RefreshToken,
		RefreshExpiry: refreshExpiry,
	}, nil
}

func tokenExpiry(token string) (time.Time, error) {
	var expiry time.Time

	parsed, err := jwt.ParseSigned(token)
	if err != nil {
		return expiry, err
	}

	var claims map[string]interface{}
	if err := parsed.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return expiry, err
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return expiry, fmt.Errorf("token has no exp claim")
	}

	return time.Unix(int64(exp), 0), nil
}

// promptForPasteCode reads the code the user copied from the redirect page.
func promptForPasteCode(in io.Reader) (string, error) {
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("failed to read the code: %v", err)
	}
	return strings.TrimSpace(line), nil
}

// requireInteractiveTerminal reports a clear error when nobody can paste a
// code, instead of waiting on a closed input forever.
func requireInteractiveTerminal() error {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	return fmt.Errorf("OAuth login needs an interactive terminal, because you must paste a code from the browser. Use a personal access token for a non-interactive session")
}

// OAuthLogin runs the OAuth 2.0 authorization code flow with PKCE. The browser
// sends the code to a static SailPoint page, the user copies it back, and this
// CLI exchanges it directly with the tenant.
func OAuthLogin() (TokenSet, error) {
	var set TokenSet

	if err := requireInteractiveTerminal(); err != nil {
		return set, err
	}

	redirect, err := redirectURI()
	if err != nil {
		return set, err
	}

	baseParsed, err := assertHTTPSURL(strings.TrimSuffix(GetBaseUrl(), "/"), "tenant API URL")
	if err != nil {
		return set, err
	}
	baseURL := baseParsed.Scheme + "://" + baseParsed.Host
	tokenEndpoint := baseURL + "/oauth/token"

	authorizeEndpoint, err := discoverAuthorizeEndpoint(baseURL)
	if err != nil {
		return set, err
	}

	codeVerifier, err := randomURLSafeString(32)
	if err != nil {
		return set, err
	}

	state, err := randomURLSafeString(32)
	if err != nil {
		return set, err
	}

	authorizeURL, err := url.Parse(authorizeEndpoint)
	if err != nil {
		return set, fmt.Errorf("authorize endpoint is not a valid URL: %v", err)
	}
	query := authorizeURL.Query()
	query.Set("client_id", ClientID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", redirect)
	query.Set("state", state)
	query.Set("code_challenge", codeChallenge(codeVerifier))
	query.Set("code_challenge_method", "S256")
	authorizeURL.RawQuery = query.Encode()

	log.Info("Opening the browser to sign in", "tenant", baseParsed.Host)
	if err := open.Run(authorizeURL.String()); err != nil {
		log.Warn("Cannot open the browser automatically", "error", err)
	}

	// Always print the URL. The browser can fail to open on a headless host, on
	// a remote shell, or inside a container.
	fmt.Fprintln(os.Stderr, "\nIf the browser did not open, go to this URL to sign in:")
	fmt.Fprintln(os.Stderr, authorizeURL.String())

	fmt.Fprintf(os.Stderr, "\nConfirmation code: %s\n", confirmationCodeFromState(state))
	fmt.Fprintln(os.Stderr, "Sign in, make sure that the page shows the same confirmation code, and copy the one-time code.")
	fmt.Fprint(os.Stderr, "\nPaste the one-time code here: ")

	pasted, err := promptForPasteCode(os.Stdin)
	if err != nil {
		return set, err
	}

	authorizationCode, err := parsePasteCode(pasted, state)
	if err != nil {
		return set, err
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", authorizationCode)
	form.Set("redirect_uri", redirect)
	form.Set("code_verifier", codeVerifier)

	response, err := requestToken(tokenEndpoint, form)
	if err != nil {
		return set, err
	}

	set, err = tokenSetFromResponse(response)
	if err != nil {
		return set, err
	}

	log.Info("OAuth authentication successful")
	return set, nil
}

// RefreshOAuth exchanges the stored refresh token with the tenant directly.
func RefreshOAuth() (TokenSet, error) {
	var set TokenSet

	refreshToken, err := GetRefreshToken()
	if err != nil {
		return set, err
	}

	baseParsed, err := assertHTTPSURL(strings.TrimSuffix(GetBaseUrl(), "/"), "tenant API URL")
	if err != nil {
		return set, err
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	response, err := requestToken(baseParsed.Scheme+"://"+baseParsed.Host+"/oauth/token", form)
	if err != nil {
		return set, err
	}

	set, err = tokenSetFromResponse(response)
	if err != nil {
		return set, err
	}

	log.Debug("OAuth token refresh successful")
	return set, nil
}
