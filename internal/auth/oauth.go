package auth

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
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/keyring"
	"github.com/sailpoint-oss/sailpoint-cli/internal/redact"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/square/go-jose.v2/jwt"
)

const (
	oauthAccessTokenService   = "environments.oauth.accesstoken"
	oauthExpiryService        = "environments.oauth.expiry"
	oauthRefreshTokenService  = "environments.oauth.refreshtoken"
	oauthRefreshExpiryService = "environments.oauth.refreshexpiry"

	OAuthClientID        = "sailpoint-cli"
	AuthLambdaBaseURL    = "https://nug87yusrg.execute-api.us-east-1.amazonaws.com/Prod/sailapps"
	AuthLambdaAuthURL    = AuthLambdaBaseURL + "/auth"
	AuthLambdaTokenURL   = AuthLambdaBaseURL + "/auth/token"
	AuthLambdaRefreshURL = AuthLambdaBaseURL + "/auth/refresh"
)

func confirmationCodeFromID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) < 8 {
		return strings.ToUpper(id)
	}

	suffix := id[len(id)-8:]
	return strings.ToUpper(suffix[:4] + "-" + suffix[4:])
}

func newOAuthTokenRequest(tokenURL, id, pickupSecret string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", tokenURL, id), nil)
	if err != nil {
		return nil, err
	}
	if pickupSecret != "" {
		req.Header.Set("Authorization", "Bearer "+pickupSecret)
	}
	return req, nil
}

// GetOAuthToken retrieves the cached OAuth access token from the keyring.
func GetOAuthToken(env string) (string, error) {
	return keyring.Get(oauthAccessTokenService, env)
}

// GetOAuthTokenExpiry retrieves the cached OAuth token expiry.
func GetOAuthTokenExpiry(env string) (time.Time, error) {
	return keyring.GetTime(oauthExpiryService, env)
}

// GetOAuthRefreshExpiry retrieves the cached OAuth refresh token expiry.
func GetOAuthRefreshExpiry(env string) (time.Time, error) {
	return keyring.GetTime(oauthRefreshExpiryService, env)
}

// GetRefreshToken retrieves the cached OAuth refresh token.
func GetRefreshToken(env string) (string, error) {
	return keyring.Get(oauthRefreshTokenService, env)
}

// CacheOAuth stores the full OAuth token set in the keyring.
func CacheOAuth(env string, set TokenSet) error {
	if err := keyring.Set(oauthAccessTokenService, env, set.AccessToken); err != nil {
		return err
	}
	if err := keyring.SetTime(oauthExpiryService, env, set.AccessExpiry); err != nil {
		return err
	}
	if err := keyring.Set(oauthRefreshTokenService, env, set.RefreshToken); err != nil {
		return err
	}
	return keyring.SetTime(oauthRefreshExpiryService, env, set.RefreshExpiry)
}

// ResetCacheOAuth clears all cached OAuth tokens for an environment.
func ResetCacheOAuth(env string) error {
	for _, svc := range []string{
		oauthAccessTokenService,
		oauthExpiryService,
		oauthRefreshTokenService,
		oauthRefreshExpiryService,
	} {
		if err := keyring.Delete(svc, env); err != nil {
			log.Debug("Failed to delete keyring entry", "service", svc, "env", env, "error", err)
		}
	}
	return nil
}

// DeleteAllOAuthSecrets removes all OAuth-related keyring entries for an environment.
func DeleteAllOAuthSecrets(env string) {
	_ = keyring.Delete(oauthAccessTokenService, env)
	_ = keyring.Delete(oauthExpiryService, env)
	_ = keyring.Delete(oauthRefreshTokenService, env)
	_ = keyring.Delete(oauthRefreshExpiryService, env)
}

// OAuthLogin performs the full OAuth browser-based authentication flow.
// baseURL is the SailPoint API base URL for this environment.
// Returns the token set and potentially an updated base URL.
func OAuthLogin(baseURL string) (TokenSet, error) {
	var set TokenSet

	privateKey, publicKeyBase64, err := generateKeyPair()
	if err != nil {
		return set, fmt.Errorf("failed to generate key pair: %v", err)
	}
	log.Debug("Generated RSA key pair for OAuth authentication")

	authRequest := AuthRequest{
		APIBaseURL: baseURL,
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
		return set, fmt.Errorf("auth lambda returned non-200 status: %d, body: %s", resp.StatusCode, redact.Bytes(bodyBytes))
	}

	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return set, fmt.Errorf("failed to decode auth lambda response: %v", err)
	}

	log.Debug("Auth response received", "id", authResponse.ID, "baseURL", authResponse.BaseURL)

	// Track if the server redirected us to a different base URL
	if authResponse.BaseURL != "" {
		set.BaseURL = authResponse.BaseURL
	}

	log.Info("Attempting to open browser for authentication")
	if err := open.Run(authResponse.AuthURL); err != nil {
		log.Warn("Cannot open automatically, Please manually open OAuth login page below")
		fmt.Println(authResponse.AuthURL)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-timeout:
			return set, fmt.Errorf("authentication timed out after 5 minutes")
		case <-ticker.C:
			tokenReq, err := newOAuthTokenRequest(AuthLambdaTokenURL, authResponse.ID, authResponse.PickupSecret)
			if err != nil {
				return set, fmt.Errorf("failed to create token polling request: %v", err)
			}

			tokenResp, err := http.DefaultClient.Do(tokenReq)
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

				if tokenResponse.BaseURL != "" {
					set.BaseURL = tokenResponse.BaseURL
				}

				var encryptedTokenData EncryptedTokenData
				if err := json.Unmarshal([]byte(tokenResponse.TokenInfo), &encryptedTokenData); err != nil {
					return set, fmt.Errorf("failed to parse encrypted token data: %v", err)
				}

				decryptedTokenInfo, err := decryptHybridToken(&encryptedTokenData, privateKey)
				if err != nil {
					return set, fmt.Errorf("failed to decrypt token info: %v", err)
				}

				var response RefreshResponse
				if err := json.Unmarshal([]byte(decryptedTokenInfo), &response); err != nil {
					return set, fmt.Errorf("failed to parse decrypted token info: %v", err)
				}

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

				set.AccessToken = response.AccessToken
				set.AccessExpiry = time.Unix(int64(accessTokenClaims["exp"].(float64)), 0)
				set.RefreshToken = response.RefreshToken
				set.RefreshExpiry = time.Unix(int64(refreshTokenClaims["exp"].(float64)), 0)

				log.Info("OAuth authentication successful")
				return set, nil
			}
			bodyBytes, _ := io.ReadAll(tokenResp.Body)
			if tokenResp.StatusCode == http.StatusUnauthorized {
				tokenResp.Body.Close()
				return set, fmt.Errorf("token polling unauthorized: %s", redact.Bytes(bodyBytes))
			}
			log.Debug("Token not ready", "status", tokenResp.StatusCode, "body", redact.Bytes(bodyBytes))
			tokenResp.Body.Close()
		}
	}
}

// RefreshOAuth uses the refresh token to obtain new access and refresh tokens.
func RefreshOAuth(env, baseURL, tenantURL string) (TokenSet, error) {
	var set TokenSet

	refreshToken, err := GetRefreshToken(env)
	if err != nil {
		return set, err
	}

	refreshRequest := RefreshRequest{
		RefreshToken: refreshToken,
		APIBaseURL:   baseURL,
		Tenant:       tenantURL,
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
		return set, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, redact.Bytes(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return set, err
	}

	var response RefreshResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return set, err
	}

	if response.AccessToken == "" {
		return set, fmt.Errorf("no access token in refresh response")
	}
	if response.RefreshToken == "" {
		return set, fmt.Errorf("no refresh token in refresh response")
	}

	var accessTokenClaims map[string]interface{}
	accToken, err := jwt.ParseSigned(response.AccessToken)
	if err != nil {
		return set, err
	}
	accToken.UnsafeClaimsWithoutVerification(&accessTokenClaims)

	var refreshTokenClaims map[string]interface{}
	refToken, err := jwt.ParseSigned(response.RefreshToken)
	if err != nil {
		return set, err
	}
	refToken.UnsafeClaimsWithoutVerification(&refreshTokenClaims)

	set = TokenSet{
		AccessToken:   response.AccessToken,
		AccessExpiry:  time.Unix(int64(accessTokenClaims["exp"].(float64)), 0),
		RefreshToken:  response.RefreshToken,
		RefreshExpiry: time.Unix(int64(refreshTokenClaims["exp"].(float64)), 0),
	}

	log.Debug("OAuth token refresh successful")
	return set, nil
}

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

func decryptHybridToken(encryptedData *EncryptedTokenData, privateKey *rsa.PrivateKey) (string, error) {
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

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return "", fmt.Errorf("RSA decryption failed: %v", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	ciphertextWithTag := append(ciphertext, authTag...)

	plaintext, err := gcm.Open(nil, iv, ciphertextWithTag, nil)
	if err != nil {
		return "", fmt.Errorf("AES-GCM decryption failed: %v", err)
	}

	return string(plaintext), nil
}
