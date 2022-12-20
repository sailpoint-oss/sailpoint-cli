package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/types"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
)

var (
	callbackErr error
	conf        *oauth2.Config
	ctx         context.Context
	server      *http.Server
	orgConfig   types.OrgConfig
	accessToken string
	expiry      time.Time
)

func callbackHandler(w http.ResponseWriter, r *http.Request) {
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

	resp, err := client.Get(orgConfig.OAuth.BaseUrl + "/beta/tenant-data/hosting-data")
	if err != nil {
		callbackErr = err
	} else {
		color.Green("Authentication successful")
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
	}

	accessToken = tok.AccessToken
	expiry = tok.Expiry

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

func OAuthLogin(config types.OrgConfig) (types.Token, error) {
	var token types.Token
	ctx = context.Background()
	conf = &oauth2.Config{
		ClientID:     config.OAuth.ClientID,
		ClientSecret: config.OAuth.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.OAuth.AuthUrl,
			TokenURL: config.OAuth.TokenUrl,
		},
		RedirectURL: "http://localhost:" + fmt.Sprint(config.OAuth.Redirect.Port) + config.OAuth.Redirect.Path,
	}

	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslClient := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslClient)

	// Redirect user to login page
	url := conf.AuthCodeURL("")

	orgConfig = config

	color.Green("Opening browser for authentication")

	open.Run(url)

	http.HandleFunc(config.OAuth.Redirect.Path, callbackHandler)
	server = &http.Server{Addr: fmt.Sprintf(":%v", config.OAuth.Redirect.Port), Handler: nil}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.ListenAndServe()
	}()
	wg.Wait()
	if callbackErr != nil {
		return token, callbackErr
	}

	token.AccessToken = accessToken
	token.Expiry = expiry
	color.Blue("%+v", token)
	return token, nil
}

func PATLogin(config types.OrgConfig, ctx context.Context) (types.Token, error) {
	var token types.Token

	uri, err := url.Parse(config.Pat.TokenUrl)
	if err != nil {
		return token, err
	}

	query := &url.Values{}
	query.Add("grant_type", "client_credentials")
	uri.RawQuery = query.Encode()

	data := &url.Values{}
	data.Add("client_id", config.Pat.ClientID)
	data.Add("client_secret", config.Pat.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return token, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return token, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return token, fmt.Errorf("failed to retrieve access token. status %s", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return token, err
	}

	var tResponse types.TokenResponse

	err = json.Unmarshal(raw, &tResponse)
	if err != nil {
		return token, err
	}

	now := time.Now()

	token.AccessToken = tResponse.AccessToken
	token.Expiry = now.Add(time.Second * time.Duration(tResponse.ExpiresIn))

	return token, nil
}
