package config

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/skratchdot/open-golang/open"
	keyring "github.com/zalando/go-keyring"
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

type TokenSet struct {
	AccessToken   string
	AccessExpiry  time.Time
	RefreshToken  string
	RefreshExpiry time.Time
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
	callbackErr error
	conf        *oauth2.Config
	ctx         context.Context
	server      *http.Server
	tokenSet    TokenSet
)

const (
	ClientID     = "sailpoint-cli"
	RedirectPort = "3000"
	RedirectPath = "/callback"
	RedirectURL  = "http://localhost:" + RedirectPort + RedirectPath
)

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

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	queryParts, _ := url.ParseQuery(r.URL.RawQuery)

	// Use the authorization code that is pushed to the redirect URL
	code := queryParts["code"][0]

	// Exchange will do the handshake to retrieve the initial access token.
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Error(err)
		callbackErr = err
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
		log.Info("Authentication successful")
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
	}

	var accessToken map[string]interface{}
	accToken, err := jwt.ParseSigned(tok.AccessToken)
	if err != nil {
		log.Error(err)
		callbackErr = err
	}
	accToken.UnsafeClaimsWithoutVerification(&accessToken)

	var refreshToken map[string]interface{}
	refToken, err := jwt.ParseSigned(tok.Extra("refresh_token").(string))
	if err != nil {
		log.Error(err)
		callbackErr = err
	}
	refToken.UnsafeClaimsWithoutVerification(&refreshToken)

	tokenSet = TokenSet{AccessToken: tok.AccessToken, AccessExpiry: time.Unix(int64(accessToken["exp"].(float64)), 0), RefreshToken: tok.Extra("refresh_token").(string), RefreshExpiry: time.Unix(int64(refreshToken["exp"].(float64)), 0)}

	// show succes page
	fmt.Fprint(w, OAuthSuccessPage)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		callbackErr = err
	}
}

func OAuthLogin() (TokenSet, error) {
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

	log.Info("Attempting to open browser for authentication")

	err := open.Run(url)
	if err != nil {
		log.Warn("Cannot open automatically, Please manually open OAuth login page below")
		fmt.Println(url)
	}

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
		return tokenSet, callbackErr
	}

	return tokenSet, nil
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
