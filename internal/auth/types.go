package auth

import "time"

// PATSet holds the result of a PAT authentication.
type PATSet struct {
	AccessToken  string
	AccessExpiry time.Time
}

// TokenSet holds the result of an OAuth authentication (access + refresh).
type TokenSet struct {
	AccessToken   string
	AccessExpiry  time.Time
	RefreshToken  string
	RefreshExpiry time.Time
	BaseURL       string // May be updated by the OAuth flow
}

// TokenResponse is the response from the /oauth/token endpoint (PAT flow).
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// RefreshResponse is the response from the OAuth refresh flow.
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

// AuthRequest represents the request body for initiating OAuth authentication.
type AuthRequest struct {
	Tenant     string `json:"tenant,omitempty"`
	APIBaseURL string `json:"apiBaseURL,omitempty"`
	PublicKey  string `json:"publicKey"`
}

// AuthResponse represents the response from the authentication initiation endpoint.
type AuthResponse struct {
	AuthURL string `json:"authURL"`
	ID      string `json:"id"`
	BaseURL string `json:"baseURL"`
	TTL     int64  `json:"ttl"`
}

// OAuthTokenResponse represents the response containing the encrypted token.
type OAuthTokenResponse struct {
	ID        string `json:"id"`
	BaseURL   string `json:"baseURL"`
	TokenInfo string `json:"tokenInfo"`
}

// EncryptedTokenData represents the structure of the encrypted token JSON.
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

// RefreshRequest represents the request body for refreshing OAuth tokens.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
	APIBaseURL   string `json:"apiBaseURL,omitempty"`
	Tenant       string `json:"tenant,omitempty"`
}

// PatConfig holds the per-environment PAT configuration (kept for mapstructure compat).
type PatConfig struct {
	ClientID     string    `mapstructure:"clientid"`
	ClientSecret string    `mapstructure:"clientsecret"`
	AccessToken  string    `mapstructure:"accesstoken"`
	Expiry       time.Time `mapstructure:"expiry"`
}
