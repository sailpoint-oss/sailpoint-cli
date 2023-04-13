package config

import (
	"context"
	"crypto/tls"
	"encoding/json"
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

func GetRefreshToken() string {
	return viper.GetString("environments." + GetActiveEnvironment() + ".oauth.refreshtoken")
}

func SetRefreshToken(token string) {
	viper.Set("environments."+GetActiveEnvironment()+".oauth.refreshtoken", token)
}

func GetOAuthRefreshExpiry() time.Time {
	return viper.GetTime("environments." + GetActiveEnvironment() + ".oauth.refreshexpiry")
}

func SetOAuthRefreshExpiry(expiry time.Time) {
	viper.Set("environments."+GetActiveEnvironment()+".oauth.refreshexpiry", expiry)
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

	clientTok := tok
	clientTok.RefreshToken = ""

	// The HTTP Client returned by conf.Client will refresh the token as necessary.
	client := conf.Client(ctx, clientTok)

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

	var accessToken map[string]interface{}
	accToken, err := jwt.ParseSigned(tok.AccessToken)
	if err != nil {
		callbackErr = err
	}
	accToken.UnsafeClaimsWithoutVerification(&accessToken)

	var refreshToken map[string]interface{}
	refToken, err := jwt.ParseSigned(tok.Extra("refresh_token").(string))
	if err != nil {
		callbackErr = err
	}
	refToken.UnsafeClaimsWithoutVerification(&refreshToken)

	SetOAuthToken(tok.AccessToken)
	SetOAuthTokenExpiry(time.Unix(int64(accessToken["exp"].(float64)), 0))

	SetRefreshToken(tok.Extra("refresh_token").(string))
	SetOAuthRefreshExpiry(time.Unix(int64(refreshToken["exp"].(float64)), 0))

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

func RefreshOAuth() error {
	var response RefreshResponse

	resp, err := http.Post(GetTokenUrl()+"?grant_type=refresh_token&client_id="+ClientID+"&refresh_token="+GetRefreshToken(), "application/json", nil)
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	var accessToken map[string]interface{}
	accToken, err := jwt.ParseSigned(response.AccessToken)
	if err != nil {
		return err
	}
	accToken.UnsafeClaimsWithoutVerification(&accessToken)

	var refreshToken map[string]interface{}
	refToken, err := jwt.ParseSigned(response.RefreshToken)
	if err != nil {
		return err
	}
	refToken.UnsafeClaimsWithoutVerification(&refreshToken)

	SetOAuthToken(response.AccessToken)
	SetOAuthTokenExpiry(time.Unix(int64(accessToken["exp"].(float64)), 0))

	SetRefreshToken(response.RefreshToken)
	SetOAuthRefreshExpiry(time.Unix(int64(refreshToken["exp"].(float64)), 0))

	return nil
}
