package config

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/kr/pretty"
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
	ClientID = "sailpoint-cli"
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

func decryptTokenInfo(encryptedToken string, encryptionKey string) (string, error) {
	// Split the IV and encrypted data
	parts := strings.Split(encryptedToken, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted token format")
	}

	// Convert hex-encoded IV and encrypted data to bytes
	iv, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %v", err)
	}

	encryptedData, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	// Convert hex-encoded encryption key to bytes
	key, err := hex.DecodeString(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode encryption key: %v", err)
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Create CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the data
	plaintext := make([]byte, len(encryptedData))
	mode.CryptBlocks(plaintext, encryptedData)

	// Remove PKCS7 padding
	paddingLen := int(plaintext[len(plaintext)-1])
	if paddingLen > aes.BlockSize || paddingLen == 0 {
		return "", fmt.Errorf("invalid padding size")
	}
	plaintext = plaintext[:len(plaintext)-paddingLen]

	return string(plaintext), nil
}

func OAuthLogin() (TokenSet, error) {
	var set TokenSet

	// Step 1: Request UUID, encryption key, and Auth URL from Auth-Lambda
	authLambdaURL := "https://nug87yusrg.execute-api.us-east-1.amazonaws.com/Prod/sailapps/uuid"

	body := bytes.NewBuffer([]byte(`{"apiBaseURL": "` + GetBaseUrl() + `"}`))

	resp, err := http.Post(authLambdaURL, "application/json", body)
	if err != nil {
		return set, fmt.Errorf("failed to get auth URL from lambda: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return set, fmt.Errorf("auth lambda returned non-200 status: %d", resp.StatusCode)
	}

	var authResponse struct {
		ID            string `json:"id"`
		EncryptionKey string `json:"encryptionKey"`
		AuthURL       string `json:"authURL"`
		BaseURL       string `json:"baseURL"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return set, fmt.Errorf("failed to decode auth lambda response: %v", err)
	}

	pretty.Print(authResponse)

	// Update the base URL for this session
	if authResponse.BaseURL != "" {
		SetBaseUrl(authResponse.BaseURL)
	}

	// Step 2: Present Auth URL to user
	log.Info("Attempting to open browser for authentication")
	err = open.Run(authResponse.AuthURL)
	if err != nil {
		log.Warn("Cannot open automatically, Please manually open OAuth login page below")
		fmt.Println(authResponse.AuthURL)
	}

	// Step 3: Poll Auth-Lambda for token using UUID
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-timeout:
			return set, fmt.Errorf("authentication timed out after 5 minutes")
		case <-ticker.C:
			// Query Auth-Lambda for token using UUID
			tokenResp, err := http.Get(fmt.Sprintf("%s/%s", authLambdaURL, authResponse.ID))
			if err != nil {
				continue
			}
			defer tokenResp.Body.Close()

			if tokenResp.StatusCode == http.StatusOK {
				var tokenResponse struct {
					BaseURL   string `json:"baseURL"`
					ID        string `json:"id"`
					TokenInfo string `json:"tokenInfo"`
				}

				if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResponse); err != nil {
					return set, fmt.Errorf("failed to decode token response: %v", err)
				}

				// Update base URL if provided in token response
				if tokenResponse.BaseURL != "" {
					SetBaseUrl(tokenResponse.BaseURL)
				}

				// Decrypt the token info using the encryption key
				decryptedTokenInfo, err := decryptTokenInfo(tokenResponse.TokenInfo, authResponse.EncryptionKey)
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

				return set, nil
			}
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

	resp, err := http.Post(GetTokenUrl()+"?grant_type=refresh_token&client_id="+ClientID+"&refresh_token="+tempRefreshToken, "application/json", nil)
	if err != nil {
		return set, err
	}
	//We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return set, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return set, err
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

	set = TokenSet{AccessToken: response.AccessToken, AccessExpiry: time.Unix(int64(accessToken["exp"].(float64)), 0), RefreshToken: response.RefreshToken, RefreshExpiry: time.Unix(int64(refreshToken["exp"].(float64)), 0)}

	return set, nil
}
