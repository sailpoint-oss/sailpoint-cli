package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func GetOAuthToken() string {
	return viper.GetString("environments." + GetActiveEnvironment() + ".oauth.accesstoken")
}

func SetOAuthToken(token string) {
	viper.Set("environments."+GetActiveEnvironment()+".oauth.accesstoken", token)
}

func GetOAuthTokenExpiry() time.Time {
	return viper.GetTime("environments." + GetActiveEnvironment() + ".oauth.expiry")
}

func SetOAuthTokenExpiry(expiry time.Time) {
	viper.Set("environments."+GetActiveEnvironment()+".oauth.expiry", expiry)
}

var (
	callbackErr error
	conf        *oauth2.Config
	ctx         context.Context
	server      *http.Server
)

const (
	ClientID     = "sailpoint-cli"
	RedirectPort = "3000"
	RedirectPath = "/callback"
	RedirectURL  = "http://localhost:" + RedirectPort + RedirectPath
)

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
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

	baseUrl := GetBaseUrl()

	resp, err := client.Get(baseUrl + "/beta/tenant-data/hosting-data")
	if err != nil {
		callbackErr = err
	} else {
		color.Green("Authentication successful")
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
	}

	SetOAuthToken(tok.AccessToken)
	SetOAuthTokenExpiry(tok.Expiry)

	// show succes page
	fmt.Fprint(w, OAuthSuccessPage)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		callbackErr = err
	}
}

func OAuthLogin() error {
	ctx = context.Background()

	conf = &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			AuthURL:  GetAuthorizeUrl(),
			TokenURL: GetTokenUrl(),
		},
		RedirectURL: RedirectURL,
	}

	selfSignedCertificateTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslClient := &http.Client{Transport: selfSignedCertificateTransport}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslClient)

	// Redirect user to login page
	url := conf.AuthCodeURL("")

	color.Green("Opening browser for authentication")

	open.Run(url)

	http.HandleFunc(RedirectPath, CallbackHandler)
	server = &http.Server{Addr: fmt.Sprintf(":%v", RedirectPort), Handler: nil}
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
