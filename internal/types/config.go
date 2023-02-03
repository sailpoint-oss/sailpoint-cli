package types

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	clierrors "github.com/sailpoint-oss/sailpoint-cli/internal/errors"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type Bundle struct {
	Config    CLIConfig
	Client    any
	APIClient *sailpoint.APIClient
}

type CLIConfig struct {
	CustomExportTemplatesPath string `mapstructure:"customExportTemplatesPath"`
	CustomSearchTemplatesPath string `mapstructure:"customSearchTemplatesPath"`
	Debug                     bool   `mapstructure:"debug"`
	AuthType                  string `mapstructure:"authtype"`
	ActiveEnvironment         string `mapstructure:"activeEnv"`
	Environments              map[string]Environment
}

func (config CLIConfig) GetAuthType() string {
	return strings.ToLower(config.AuthType)
}

func (config CLIConfig) GetBaseUrl() (string, error) {
	switch config.GetAuthType() {
	case "pat":
		return config.Environments[config.ActiveEnvironment].Pat.BaseUrl, nil
	case "oauth":
		return config.Environments[config.ActiveEnvironment].OAuth.BaseUrl, nil
	default:
		return "", fmt.Errorf("configured authtype ('%s') is invalid or missing", config.AuthType)
	}
}

var (
	callbackErr error
	conf        *oauth2.Config
	ctx         context.Context
	server      *http.Server
)

func (config CLIConfig) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	queryParts, _ := url.ParseQuery(r.URL.RawQuery)

	// Use the authorization code that is pushed to the redirect URL
	code := queryParts["code"][0]

	// Exchange will do the handshake to retrieve the initial access token.
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	// The HTTP Client returned by conf.Client will refresh the token as necessary.
	client := conf.Client(ctx, tok)

	baseUrl, err := config.GetBaseUrl()
	if err != nil {
		callbackErr = err
	}

	resp, err := client.Get(baseUrl + "/beta/tenant-data/hosting-data")
	if err != nil {
		callbackErr = err
	} else {
		color.Green("Authentication successful")
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
	}

	viper.Set(fmt.Sprintf("environments.%s.oauth.token", config.ActiveEnvironment), Token{AccessToken: tok.AccessToken, Expiry: tok.Expiry})

	config.SaveConfig()

	// show succes page
	msg := "<p><strong>SailPoint CLI, OAuth Login Success!</strong></p>"
	msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
	fmt.Fprint(w, msg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		callbackErr = err
	}
}

func (config CLIConfig) OAuthLogin() error {
	ctx = context.Background()

	conf = &oauth2.Config{
		ClientID:     config.Environments[config.ActiveEnvironment].OAuth.ClientID,
		ClientSecret: config.Environments[config.ActiveEnvironment].OAuth.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.Environments[config.ActiveEnvironment].OAuth.AuthUrl,
			TokenURL: config.Environments[config.ActiveEnvironment].OAuth.TokenUrl,
		},
		RedirectURL: "http://localhost:" + fmt.Sprint(config.Environments[config.ActiveEnvironment].OAuth.Redirect.Port) + config.Environments[config.ActiveEnvironment].OAuth.Redirect.Path,
	}

	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslClient := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslClient)

	// Redirect user to login page
	url := conf.AuthCodeURL("")

	color.Green("Opening browser for authentication")

	open.Run(url)

	http.HandleFunc(config.Environments[config.ActiveEnvironment].OAuth.Redirect.Path, config.CallbackHandler)
	server = &http.Server{Addr: fmt.Sprintf(":%v", config.Environments[config.ActiveEnvironment].OAuth.Redirect.Port), Handler: nil}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.ListenAndServe()
	}()
	wg.Wait()
	if callbackErr != nil {
		return callbackErr
	}

	return nil
}

func (config CLIConfig) PATLogin() error {

	uri, err := url.Parse(config.Environments[config.ActiveEnvironment].Pat.TokenUrl)
	if err != nil {
		return err
	}

	query := &url.Values{}
	query.Add("grant_type", "client_credentials")
	uri.RawQuery = query.Encode()

	data := &url.Values{}
	data.Add("client_id", config.Environments[config.ActiveEnvironment].Pat.ClientID)
	data.Add("client_secret", config.Environments[config.ActiveEnvironment].Pat.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to retrieve access token. status %s", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tResponse TokenResponse

	err = json.Unmarshal(raw, &tResponse)
	if err != nil {
		return err
	}

	now := time.Now()

	viper.Set(fmt.Sprintf("environments.%s.pat.token", config.ActiveEnvironment), Token{AccessToken: tResponse.AccessToken, Expiry: now.Add(time.Second * time.Duration(tResponse.ExpiresIn))})

	config.SaveConfig()

	return nil
}

func (config CLIConfig) EnsureAccessToken() error {
	err := config.Validate()
	if err != nil {
		return err
	}

	authType := config.GetAuthType()

	switch authType {
	case "pat":
		if config.Environments[config.ActiveEnvironment].Pat.Token.Expiry.After(time.Now()) {
			return nil
		}
	case "oauth":
		if config.Environments[config.ActiveEnvironment].OAuth.Token.Expiry.After(time.Now()) {
			return nil
		}
	default:
		return errors.New("invalid authtype configured")
	}

	switch authType {
	case "pat":
		err = config.PATLogin()
		if err != nil {
			return err
		}

	case "oauth":
		err = config.OAuthLogin()
		if err != nil {
			return err
		}

	default:
		return errors.New("invalid authtype configured")
	}

	config.SaveConfig()

	return nil

}

func (config CLIConfig) SaveConfig() error {
	err := viper.WriteConfig()
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

func (config CLIConfig) GetAuthToken() (string, error) {
	authType := config.GetAuthType()
	switch authType {
	case "pat":
		expiry := viper.GetTime(fmt.Sprintf("environments.%s.pat.token.expiry", config.ActiveEnvironment))
		fmt.Println(expiry)
		if expiry.After(time.Now()) {
			return viper.GetString(fmt.Sprintf("environments.%s.pat.token.accesstoken", config.ActiveEnvironment)), nil
		} else {
			return "", clierrors.ErrAccessTokenExpired
		}
	case "oauth":
		expiry := viper.GetTime(fmt.Sprintf("environments.%s.oauth.token.expiry", config.ActiveEnvironment))
		fmt.Println(expiry)
		if expiry.After(time.Now()) {
			return viper.GetString(fmt.Sprintf("environments.%s.oauth.token.accesstoken", config.ActiveEnvironment)), nil
		} else {
			return "", clierrors.ErrAccessTokenExpired
		}
	default:
		return "", fmt.Errorf("invalid authtype '%s' configured", config.AuthType)

	}
}

func (config CLIConfig) GetActiveEnvironment(env string) Environment {
	return config.Environments[config.ActiveEnvironment]
}

func (config CLIConfig) GetEnvironment(env string) Environment {
	return config.Environments[env]
}

func (config CLIConfig) Validate() error {
	switch config.GetAuthType() {
	case "pat":
		if config.Environments[config.ActiveEnvironment].Pat.TokenUrl == "" {
			return fmt.Errorf("missing PAT TokenURL configuration value")
		}
		if config.Environments[config.ActiveEnvironment].Pat.ClientID == "" {
			return fmt.Errorf("missing PAT ClientID configuration value")
		}
		if config.Environments[config.ActiveEnvironment].Pat.ClientSecret == "" {
			return fmt.Errorf("missing PAT ClientSecret configuration value")
		}
		return nil
	case "oauth":
		if config.Environments[config.ActiveEnvironment].OAuth.AuthUrl == "" {
			return fmt.Errorf("missing OAuth URL configuration value")
		}
		if config.Environments[config.ActiveEnvironment].OAuth.ClientID == "" {
			return fmt.Errorf("missing OAuth ClientID configuration value")
		}
		if config.Environments[config.ActiveEnvironment].OAuth.ClientSecret == "" && config.Debug {
			color.Yellow("missing OAuth ClientSecret configuration value")
		}
		if config.Environments[config.ActiveEnvironment].OAuth.Redirect.Path == "" {
			return fmt.Errorf("missing OAuth Redirect Path configuration value")
		}
		if config.Environments[config.ActiveEnvironment].OAuth.Redirect.Port == 0 {
			return fmt.Errorf("missing OAuth Redirect Port configuration value")
		}
		if config.Environments[config.ActiveEnvironment].OAuth.TokenUrl == "" {
			return fmt.Errorf("missing OAuth TokenUrl configuration value")
		}
		return nil
	default:
		return fmt.Errorf("invalid authtype '%s' configured", config.AuthType)
	}
}
