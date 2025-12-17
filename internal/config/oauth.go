package config

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/skratchdot/open-golang/open"
	keyring "github.com/zalando/go-keyring"
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

// AuthRequest represents the request body for initiating OAuth authentication
type AuthRequest struct {
	Tenant     string `json:"tenant,omitempty"`
	APIBaseURL string `json:"apiBaseURL,omitempty"`
	PublicKey  string `json:"publicKey"`
}

// AuthResponse represents the response from the authentication initiation endpoint
type AuthResponse struct {
	AuthURL string `json:"authURL"`
	ID      string `json:"id"`
	BaseURL string `json:"baseURL"`
	TTL     int64  `json:"ttl"`
}

// OAuthTokenResponse represents the response containing the encrypted token from OAuth flow
type OAuthTokenResponse struct {
	ID        string `json:"id"`
	BaseURL   string `json:"baseURL"`
	TokenInfo string `json:"tokenInfo"`
}

// EncryptedTokenData represents the structure of the encrypted token JSON
type EncryptedTokenData struct {
	Version   string `json:"version"`
	Algorithm struct {
		Symmetric  string `json:"symmetric"`
		Asymmetric string `json:"asymmetric"`
	} `json:"algorithm"`
	Data struct {
		Ciphertext   string `json:"ciphertext"`
		EncryptedKey string `json:"encryptedKey"`
		IV           string `json:"iv"`
		AuthTag      string `json:"authTag"`
	} `json:"data"`
}

// RefreshRequest represents the request body for refreshing OAuth tokens
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
	APIBaseURL   string `json:"apiBaseURL,omitempty"`
	Tenant       string `json:"tenant,omitempty"`
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

var (
	tokenSet TokenSet
)

const (
	ClientID             = "sailpoint-cli"
	AuthLambdaBaseURL    = "https://nug87yusrg.execute-api.us-east-1.amazonaws.com/Prod/sailapps"
	AuthLambdaAuthURL    = AuthLambdaBaseURL + "/auth"
	AuthLambdaTokenURL   = AuthLambdaBaseURL + "/auth/token"
	AuthLambdaRefreshURL = AuthLambdaBaseURL + "/auth/refresh"
)

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

// generateKeyPair creates a new 2048-bit RSA key pair for OAuth authentication
// Returns the private key, the public key as base64-encoded PEM, and any error
func generateKeyPair() (*rsa.PrivateKey, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyPEM)
	return privateKey, publicKeyBase64, nil
}

// decryptHybridToken decrypts a token encrypted with hybrid RSA-OAEP + AES-256-GCM encryption
func decryptHybridToken(encryptedData *EncryptedTokenData, privateKey *rsa.PrivateKey) (string, error) {
	// 1. Decode base64 components
	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedData.Data.EncryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted key: %v", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData.Data.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %v", err)
	}

	iv, err := base64.StdEncoding.DecodeString(encryptedData.Data.IV)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %v", err)
	}

	authTag, err := base64.StdEncoding.DecodeString(encryptedData.Data.AuthTag)
	if err != nil {
		return "", fmt.Errorf("failed to decode auth tag: %v", err)
	}

	// 2. Decrypt AES key using RSA-OAEP-SHA256
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return "", fmt.Errorf("RSA decryption failed: %v", err)
	}

	// 3. Decrypt token using AES-256-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// Append auth tag to ciphertext (GCM expects it this way)
	ciphertextWithTag := append(ciphertext, authTag...)

	plaintext, err := gcm.Open(nil, iv, ciphertextWithTag, nil)
	if err != nil {
		return "", fmt.Errorf("AES-GCM decryption failed: %v", err)
	}

	return string(plaintext), nil
}

func OAuthLogin() (TokenSet, error) {
	var set TokenSet

	// Step 1: Generate RSA key pair for this authentication session
	privateKey, publicKeyBase64, err := generateKeyPair()
	if err != nil {
		return set, fmt.Errorf("failed to generate key pair: %v", err)
	}
	log.Debug("Generated RSA key pair for OAuth authentication")

	// Step 2: Initiate authentication flow with the public key
	authRequest := AuthRequest{
		APIBaseURL: GetBaseUrl(),
		PublicKey:  publicKeyBase64,
	}

	requestBody, err := json.Marshal(authRequest)
	if err != nil {
		return set, fmt.Errorf("failed to marshal auth request: %v", err)
	}

	resp, err := http.Post(AuthLambdaAuthURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return set, fmt.Errorf("failed to initiate auth with lambda: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return set, fmt.Errorf("auth lambda returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return set, fmt.Errorf("failed to decode auth lambda response: %v", err)
	}

	log.Debug("Auth response received", "id", authResponse.ID, "baseURL", authResponse.BaseURL)

	// Update the base URL for this session
	if authResponse.BaseURL != "" {
		SetBaseUrl(authResponse.BaseURL)
	}

	// Step 3: Present Auth URL to user
	log.Info("Attempting to open browser for authentication")
	err = open.Run(authResponse.AuthURL)
	if err != nil {
		log.Warn("Cannot open automatically, Please manually open OAuth login page below")
		fmt.Println(authResponse.AuthURL)
	}

	// Step 4: Poll Auth-Lambda for encrypted token using UUID
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-timeout:
			return set, fmt.Errorf("authentication timed out after 5 minutes")
		case <-ticker.C:
			// Query Auth-Lambda for token using UUID
			tokenResp, err := http.Get(fmt.Sprintf("%s/%s", AuthLambdaTokenURL, authResponse.ID))
			if err != nil {
				log.Debug("Error polling for token", "error", err)
				continue
			}

			if tokenResp.StatusCode == http.StatusOK {
				var tokenResponse OAuthTokenResponse
				if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResponse); err != nil {
					tokenResp.Body.Close()
					return set, fmt.Errorf("failed to decode token response: %v", err)
				}
				tokenResp.Body.Close()

				// Update base URL if provided in token response
				if tokenResponse.BaseURL != "" {
					SetBaseUrl(tokenResponse.BaseURL)
				}

				// Parse the encrypted token data
				var encryptedTokenData EncryptedTokenData
				if err := json.Unmarshal([]byte(tokenResponse.TokenInfo), &encryptedTokenData); err != nil {
					return set, fmt.Errorf("failed to parse encrypted token data: %v", err)
				}

				// Decrypt the token using our private key
				decryptedTokenInfo, err := decryptHybridToken(&encryptedTokenData, privateKey)
				if err != nil {
					return set, fmt.Errorf("failed to decrypt token info: %v", err)
				}

				// Parse the decrypted token info into RefreshResponse
				var response RefreshResponse
				if err := json.Unmarshal([]byte(decryptedTokenInfo), &response); err != nil {
					return set, fmt.Errorf("failed to parse decrypted token info: %v", err)
				}

				// Parse tokens to get expiry
				var accessTokenClaims map[string]interface{}
				accToken, err := jwt.ParseSigned(response.AccessToken)
				if err != nil {
					return set, fmt.Errorf("failed to parse access token: %v", err)
				}
				accToken.UnsafeClaimsWithoutVerification(&accessTokenClaims)

				var refreshTokenClaims map[string]interface{}
				refToken, err := jwt.ParseSigned(response.RefreshToken)
				if err != nil {
					return set, fmt.Errorf("failed to parse refresh token: %v", err)
				}
				refToken.UnsafeClaimsWithoutVerification(&refreshTokenClaims)

				set = TokenSet{
					AccessToken:   response.AccessToken,
					AccessExpiry:  time.Unix(int64(accessTokenClaims["exp"].(float64)), 0),
					RefreshToken:  response.RefreshToken,
					RefreshExpiry: time.Unix(int64(refreshTokenClaims["exp"].(float64)), 0),
				}

				log.Info("OAuth authentication successful")
				return set, nil
			}
			tokenResp.Body.Close()
		}
	}
}

func RefreshOAuth() (TokenSet, error) {
	var response RefreshResponse
	var set TokenSet

	tempRefreshToken, err := GetRefreshToken()
	if err != nil {
		return set, err
	}

	// Prepare the refresh request body
	refreshRequest := RefreshRequest{
		RefreshToken: tempRefreshToken,
		APIBaseURL:   GetBaseUrl(),
		Tenant:       GetTenantUrl(),
	}

	requestBody, err := json.Marshal(refreshRequest)
	if err != nil {
		return set, fmt.Errorf("failed to marshal refresh request: %v", err)
	}

	resp, err := http.Post(AuthLambdaRefreshURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return set, fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return set, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return set, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return set, err
	}

	if response.AccessToken == "" {
		return set, fmt.Errorf("no access token in refresh response")
	}

	if response.RefreshToken == "" {
		return set, fmt.Errorf("no refresh token in refresh response")
	}

	var accessToken map[string]interface{}
	accToken, err := jwt.ParseSigned(response.AccessToken)
	if err != nil {
		return set, err
	}
	accToken.UnsafeClaimsWithoutVerification(&accessToken)

	var refreshToken map[string]interface{}
	refToken, err := jwt.ParseSigned(response.RefreshToken)
	if err != nil {
		return set, err
	}
	refToken.UnsafeClaimsWithoutVerification(&refreshToken)

	set = TokenSet{
		AccessToken:   response.AccessToken,
		AccessExpiry:  time.Unix(int64(accessToken["exp"].(float64)), 0),
		RefreshToken:  response.RefreshToken,
		RefreshExpiry: time.Unix(int64(refreshToken["exp"].(float64)), 0),
	}

	log.Debug("OAuth token refresh successful")
	return set, nil
}
