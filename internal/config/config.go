package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/keyring"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/viper"
	"gopkg.in/square/go-jose.v2/jwt"
)

var ErrAccessTokenExpired = fmt.Errorf("accesstoken is expired")

const (
	configFolder   = ".sailpoint"
	configYamlFile = "config.yaml"
)

// Token holds OAuth token data (kept for mapstructure compatibility with existing config files).
type Token struct {
	AccessToken string    `mapstructure:"accesstoken"`
	Expiry      time.Time `mapstructure:"expiry"`

	RefreshToken  string    `mapstructure:"refreshtoken"`
	RefreshExpiry time.Time `mapstructure:"refreshexpiry"`
}

// Environment represents a single CLI environment configuration.
type Environment struct {
	TenantURL string         `mapstructure:"tenanturl"`
	BaseURL   string         `mapstructure:"baseurl"`
	AuthType  string         `mapstructure:"authtype"`
	Pat       auth.PatConfig `mapstructure:"pat"`
	OAuth     Token          `mapstructure:"oauth"`
}

// CLIConfig is the top-level CLI configuration structure.
type CLIConfig struct {
	// Standard Variables
	ExportTemplatesPath string `mapstructure:"exporttemplatespath"`
	SearchTemplatesPath string `mapstructure:"searchtemplatespath"`
	ReportTemplatesPath string `mapstructure:"reporttemplatespath"`

	Debug             bool                   `mapstructure:"debug"`
	AuthType          string                 `mapstructure:"authtype"`
	ActiveEnvironment string                 `mapstructure:"activeenvironment"`
	Environments      map[string]Environment `mapstructure:"environments"`

	// Pipeline Variables
	ClientID     string    `mapstructure:"clientid, omitempty"`
	ClientSecret string    `mapstructure:"clientsecret, omitempty"`
	BaseURL      string    `mapstructure:"base_url, omitempty"`
	AccessToken  string    `mapstructure:"accesstoken"`
	Expiry       time.Time `mapstructure:"expiry"`
}

// --- Global settings ---

func GetCustomSearchTemplatePath() string {
	return viper.GetString("searchtemplatespath")
}

func GetCustomExportTemplatePath() string {
	return viper.GetString("exporttemplatespath")
}

func GetCustomReportTemplatePath() string {
	return viper.GetString("reporttemplatespath")
}

func SetCustomSearchTemplatePath(path string) {
	viper.Set("searchtemplatespath", path)
}

func SetCustomExportTemplatePath(path string) {
	viper.Set("exporttemplatespath", path)
}

func SetCustomReportTemplatePath(path string) {
	viper.Set("reporttemplatespath", path)
}

func GetDebug() bool {
	return viper.GetBool("debug")
}

func SetDebug(debug bool) {
	viper.Set("debug", debug)
}

func GetJSONOutput() bool {
	return viper.GetBool("json")
}

// --- Environment management ---

func GetEnvironments() map[string]interface{} {
	return viper.GetStringMap("environments")
}

func GetActiveEnvironment() string {
	env := strings.ToLower(viper.GetString("activeenvironment"))
	if env == "" {
		return "default"
	}
	return env
}

func SetActiveEnvironment(activeEnv string) {
	viper.Set("activeenvironment", strings.ToLower(activeEnv))
}

// GetAuthType returns the auth type for the active environment.
// Falls back to the global authtype for backward compatibility with old configs.
func GetAuthType() string {
	if authType := os.Getenv("SAIL_AUTHTYPE"); authType != "" {
		return strings.ToLower(authType)
	}

	env := GetActiveEnvironment()
	perEnv := viper.GetString("environments." + env + ".authtype")
	if perEnv != "" {
		return strings.ToLower(perEnv)
	}
	if authType := viper.GetString("authtype"); authType != "" {
		return strings.ToLower(authType)
	}
	return "pat"
}

// SetAuthType sets the auth type for the active environment.
func SetAuthType(authType string) {
	env := GetActiveEnvironment()
	viper.Set("environments."+env+".authtype", strings.ToLower(authType))
}

// GetEnvAuthType returns the auth type for a specific environment.
func GetEnvAuthType(env string) string {
	if authType := os.Getenv("SAIL_AUTHTYPE"); authType != "" {
		return strings.ToLower(authType)
	}

	perEnv := viper.GetString("environments." + env + ".authtype")
	if perEnv != "" {
		return strings.ToLower(perEnv)
	}
	if authType := viper.GetString("authtype"); authType != "" {
		return strings.ToLower(authType)
	}
	return "pat"
}

// SetEnvAuthType sets the auth type for a specific environment.
func SetEnvAuthType(env, authType string) {
	viper.Set("environments."+env+".authtype", strings.ToLower(authType))
}

// --- URL management ---

func GetEnvBaseUrl(env string) string {
	return viper.GetString("environments." + env + ".baseurl")
}

func GetBaseUrl() string {
	envBaseUrl := os.Getenv("SAIL_BASE_URL")
	if envBaseUrl != "" {
		return envBaseUrl
	}
	return GetEnvBaseUrl(GetActiveEnvironment())
}

func GetTenantUrl() string {
	return viper.GetString("environments." + GetActiveEnvironment() + ".tenanturl")
}

func GetEnvTenantUrl(env string) string {
	return viper.GetString("environments." + env + ".tenanturl")
}

func SetBaseUrl(baseUrl string) {
	viper.Set("environments."+GetActiveEnvironment()+".baseurl", baseUrl)
}

func SetEnvBaseUrl(env, baseUrl string) {
	viper.Set("environments."+env+".baseurl", baseUrl)
}

func SetTenantUrl(tenantUrl string) {
	viper.Set("environments."+GetActiveEnvironment()+".tenanturl", tenantUrl)
}

func SetEnvTenantUrl(env, tenantUrl string) {
	viper.Set("environments."+env+".tenanturl", tenantUrl)
}

func GetEnvTokenUrl(env string) string {
	return GetEnvBaseUrl(env) + "/oauth/token"
}

func GetTokenUrl() string {
	return GetBaseUrl() + "/oauth/token"
}

func GetAuthorizeUrl() string {
	return GetTenantUrl() + "/oauth/authorize"
}

// --- Initialization ---

func InitConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	viper.AddConfigPath(filepath.Join(home, ".sailpoint"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("sail")

	viper.SetDefault("authtype", "pat")
	viper.SetDefault("exporttemplatespath", "")
	viper.SetDefault("searchtemplatespath", "")
	viper.SetDefault("reporttemplatespath", "")
	viper.SetDefault("debug", false)
	viper.SetDefault("activeenvironment", "default")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore -- may be using env vars
		} else {
			return err
		}
	}

	if GetDebug() {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	}

	// Pre-warm the keyring support check so it's not done on every Validate
	_ = keyring.IsSupported()

	return nil
}

func GetConfig() (CLIConfig, error) {
	var cfg CLIConfig
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func SaveConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(home, configFolder)); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Join(home, configFolder), 0777)
		if err != nil {
			log.Warn("failed to create config folder", "folder", configFolder, "error", err)
		}
	}

	err = viper.WriteConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.SafeWriteConfig()
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// --- Auth wrappers (delegate to internal/auth) ---
// These maintain backward compatibility so existing callers of config.GetAuthToken()
// and config.InitAPIClient() continue to work.

func GetAuthToken() (string, error) {
	env := GetActiveEnvironment()
	authType := GetAuthType()
	baseURL := GetBaseUrl()
	tenantURL := GetTenantUrl()
	tokenURL := GetTokenUrl()

	return auth.GetToken(authType, env, baseURL, tenantURL, tokenURL, func(newBaseURL string) {
		SetBaseUrl(newBaseURL)
	})
}

func Validate() error {
	return auth.ValidateAuth(GetAuthType(), GetActiveEnvironment(), GetBaseUrl(), GetTenantUrl())
}

func CheckToken(tokenString string) error {
	return auth.CheckToken(tokenString)
}

func InitAPIClient(experimental bool) (*sailpoint.APIClient, error) {
	var apiClient *sailpoint.APIClient

	err := Validate()
	if err != nil {
		return apiClient, err
	}

	token, err := GetAuthToken()
	if err != nil {
		log.Debug("unable to retrieve accesstoken", "error", err)
	}

	configuration := sailpoint.NewCLIConfiguration(sailpoint.ClientConfiguration{Token: token, BaseURL: GetBaseUrl()})

	if experimental {
		configuration.Experimental = true
	}

	apiClient = sailpoint.NewAPIClient(configuration)
	if GetDebug() {
		logger := log.NewWithOptions(os.Stdout, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			Level:           log.DebugLevel,
		})
		debugLogger := logger.StandardLog(log.StandardLogOptions{ForceLevel: log.DebugLevel})
		apiClient.V3.GetConfig().HTTPClient.Logger = debugLogger
		apiClient.Beta.GetConfig().HTTPClient.Logger = debugLogger
	} else {
		var DevNull types.DevNull
		apiClient.V3.GetConfig().HTTPClient.Logger = DevNull
		apiClient.Beta.GetConfig().HTTPClient.Logger = DevNull
	}

	return apiClient, nil
}

// --- Legacy keyring wrappers for backward compatibility ---
// These delegate to internal/auth but use GetActiveEnvironment() for the env param,
// matching the old behavior. Existing callers (e.g. cmd/set/pat.go, cmd/environment/delete.go)
// can continue to work.

func GetPatClientID() (string, error) {
	return auth.GetPatClientID(GetActiveEnvironment())
}

func GetPatClientSecret() (string, error) {
	return auth.GetPatClientSecret(GetActiveEnvironment())
}

func SetPatClientID(clientID string) error {
	return auth.SetPatClientID(GetActiveEnvironment(), clientID)
}

func SetPatClientSecret(clientSecret string) error {
	return auth.SetPatClientSecret(GetActiveEnvironment(), clientSecret)
}

func ResetCachePAT() error {
	return auth.ResetCachePAT(GetActiveEnvironment())
}

func DeletePatToken(env string) error {
	if env == "" {
		env = GetActiveEnvironment()
	}
	return keyring.Delete("environments.pat.accesstoken", env)
}

func DeletePatTokenExpiry(env string) error {
	if env == "" {
		env = GetActiveEnvironment()
	}
	return keyring.Delete("environments.pat.expiry", env)
}

func DeletePatClientID(env string) error {
	if env == "" {
		env = GetActiveEnvironment()
	}
	return keyring.Delete("environments.pat.clientid", env)
}

func DeletePatClientSecret(env string) error {
	if env == "" {
		env = GetActiveEnvironment()
	}
	return keyring.Delete("environments.pat.clientsecret", env)
}

func DeleteOAuthToken(env string) error {
	if env == "" {
		env = GetActiveEnvironment()
	}
	return keyring.Delete("environments.oauth.accesstoken", env)
}

func DeleteOAuthTokenExpiry(env string) error {
	if env == "" {
		env = GetActiveEnvironment()
	}
	return keyring.Delete("environments.oauth.expiry", env)
}

func DeleteRefreshToken(env string) error {
	if env == "" {
		env = GetActiveEnvironment()
	}
	return keyring.Delete("environments.oauth.refreshtoken", env)
}

func DeleteRefreshTokenExpiry(env string) error {
	if env == "" {
		env = GetActiveEnvironment()
	}
	return keyring.Delete("environments.oauth.refreshexpiry", env)
}

func PromptForClientID() (string, error) {
	return auth.PromptForClientID()
}

func PromptForClientSecret() (string, error) {
	return auth.PromptForClientSecret()
}

// TestSecretsStorage delegates to the cached keyring check.
func TestSecretsStorage() bool {
	return keyring.IsSupported()
}

// SetTime formats a time.Time as RFC3339.
func SetTime(inputTime time.Time) string {
	return inputTime.Format(time.RFC3339)
}

// GetTime parses an RFC3339 string into a time.Time.
func GetTime(inputString string) (time.Time, error) {
	return time.Parse(time.RFC3339, inputString)
}

// --- Legacy functions that are no longer used by internal/auth but may be used elsewhere ---

func GetClientID(env string) (string, error) {
	return auth.GetPatClientID(env)
}

func GetClientSecret(env string) (string, error) {
	return auth.GetPatClientSecret(env)
}

// PATLogin delegates to internal/auth.
func PATLogin() (auth.PATSet, error) {
	clientID, err := GetPatClientID()
	if err != nil {
		return auth.PATSet{}, err
	}
	clientSecret, err := GetPatClientSecret()
	if err != nil {
		return auth.PATSet{}, err
	}
	return auth.PATLogin(GetTokenUrl(), clientID, clientSecret)
}

// CachePAT delegates to internal/auth.
func CachePAT(set auth.PATSet) error {
	return auth.CachePAT(GetActiveEnvironment(), set)
}

// OAuthLogin delegates to internal/auth.
func OAuthLogin() (auth.TokenSet, error) {
	return auth.OAuthLogin(GetBaseUrl())
}

// RefreshOAuth delegates to internal/auth.
func RefreshOAuth() (auth.TokenSet, error) {
	env := GetActiveEnvironment()
	return auth.RefreshOAuth(env, GetBaseUrl(), GetTenantUrl())
}

// CacheOAuth delegates to internal/auth.
func CacheOAuth(set auth.TokenSet) error {
	return auth.CacheOAuth(GetActiveEnvironment(), set)
}

// ResetCacheOAuth delegates to internal/auth.
func ResetCacheOAuth() error {
	return auth.ResetCacheOAuth(GetActiveEnvironment())
}

// CheckTokenForClaims returns JWT claims for display purposes.
func CheckTokenForClaims(tokenString string) (map[string]interface{}, error) {
	var claims map[string]interface{}
	token, err := jwt.ParseSigned(tokenString)
	if err != nil {
		return nil, err
	}
	token.UnsafeClaimsWithoutVerification(&claims)
	return claims, nil
}
