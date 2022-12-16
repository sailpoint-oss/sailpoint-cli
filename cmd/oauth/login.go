// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package oauth

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var (
	conf      *oauth2.Config
	ctx       context.Context
	server    *http.Server
	OAuthConf OAuthConfig
)

const (
	baseURLTemplate  = "https://%s.api.identitynow.com"
	tokenURLTemplate = "%s/oauth/token"
	configFolder     = ".sailpoint"
	configYamlFile   = "config.yaml"
)

type OAuthConfig struct {
	OAuthTenant      string    `mapstructure:"OAuthTenant"`
	OAuthBaseUrl     string    `mapstructure:"OAuthBaseUrl"`
	OAuthAccessToken string    `mapstructure:"OAuthAccessToken"`
	OAuthExpiry      time.Time `mapstructure:"OAuthExpiry"`
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	queryParts, _ := url.ParseQuery(r.URL.RawQuery)

	// fmt.Printf("%+v\n", r)

	// Use the authorization code that is pushed to the redirect
	// URL.
	code := queryParts["code"][0]
	// fmt.Printf("Authorization Code: %v\n", code)

	// Exchange will do the handshake to retrieve the initial access token.
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	// The HTTP Client returned by conf.Client will refresh the token as necessary.
	client := conf.Client(ctx, tok)

	baseUrl := fmt.Sprintf(baseURLTemplate, OAuthConf.OAuthTenant)

	resp, err := client.Get(baseUrl + "/beta/tenant-data/hosting-data")
	if err != nil {
		log.Fatal(err)
	} else {
		color.Green("Authentication successful")
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		// body, _ := io.ReadAll(resp.Body)
		// color.Green("%s\n", body)
	}

	OAuthConf.OAuthBaseUrl = baseUrl
	OAuthConf.OAuthAccessToken = tok.AccessToken
	OAuthConf.OAuthExpiry = tok.Expiry

	updateErr := updateOAuthConfig(OAuthConf)
	if updateErr != nil {
		log.Fatal(updateErr)
	}

	// show succes page
	msg := "<p><strong>Success!</strong></p>"
	msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
	fmt.Fprint(w, msg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func FormatAPIUrl(tenant string, path string) string {
	return fmt.Sprintf("https://%v.api.identitynow.com%v", tenant, path)
}

func updateOAuthConfig(conf OAuthConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(home, configFolder)); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Join(home, configFolder), 0777)
		if err != nil {
			log.Printf("failed to create %s folder for config. %v", configFolder, err)
		}
	}

	viper.Set("OAuthTenant", conf.OAuthTenant)
	viper.Set("OAuthBaseUrl", conf.OAuthBaseUrl)
	viper.Set("OAuthAccessToken", conf.OAuthAccessToken)
	viper.Set("OAuthExpiry", conf.OAuthExpiry)

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

func newLoginCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login",
		Short:   "login transform",
		Long:    "login to an IdentityNow tenant using OAuth.",
		Example: "sail oauth login",
		Aliases: []string{"l"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientId := cmd.Flags().Lookup("clientId").Value.String()
			clientSecret := cmd.Flags().Lookup("clientSecret").Value.String()
			callbackPort := cmd.Flags().Lookup("callbackPort").Value.String()
			strconv.Atoi(callbackPort)
			tenant := args[0]
			OAuthConf.OAuthTenant = tenant
			APIUrl := FormatAPIUrl(tenant, "")
			AuthUrl := fmt.Sprintf("https://%v.identitynow.com/oauth/authorize", tenant)
			TokenUrl := FormatAPIUrl(tenant, "/oauth/token")
			CallBackAddress := fmt.Sprintf("http://localhost:%v/callback/%v", callbackPort, tenant)

			fmt.Printf("Starting OAuth flow\nTenant: %v\nClient ID: %v\nClient Secret: %v\nCallBack Address: %v\nAPI URL: %v\nAuth URL: %v\nToken URL: %v\n\n\n\n", tenant, clientId, clientSecret, CallBackAddress, APIUrl, AuthUrl, TokenUrl)

			ctx = context.Background()
			conf = &oauth2.Config{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:  AuthUrl,
					TokenURL: TokenUrl,
				},
				RedirectURL: CallBackAddress,
			}

			// add transport for self-signed certificate to context
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			sslcli := &http.Client{Transport: tr}
			ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

			// Redirect user to consent page to ask for permission
			// for the scopes specified above.
			url := conf.AuthCodeURL("")

			color.Green("Opening browser for authentication")
			// time.Sleep(1 * time.Second)
			open.Run(url)
			// time.Sleep(1 * time.Second)
			// log.Printf("Authentication URL: %s\n", url)
			http.HandleFunc(fmt.Sprintf("/callback/%v", tenant), callbackHandler)
			server = &http.Server{Addr: fmt.Sprintf(":%v", callbackPort), Handler: nil}
			server.ListenAndServe()

			return nil
		},
	}

	cmd.Flags().StringP("clientId", "c", "", "The Client Id to use for the OAuth login")
	cmd.Flags().StringP("clientSecret", "s", "", "The Client Secret to use for the OAuth login")
	cmd.Flags().Int64P("callbackPort", "p", 3000, "The localhost Callback port to use for the OAuth login")

	return cmd
}
