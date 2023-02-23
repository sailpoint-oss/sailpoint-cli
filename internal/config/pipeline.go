package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func GetPipelineToken() string {
	return viper.GetString("accesstoken")
}

func SetPipelineToken(token string) {
	viper.Set("accesstoken", token)
}

func GetPipelineTokenExpiry() time.Time {
	return viper.GetTime("expiry")
}

func SetPipelineTokenExpiry(expiry time.Time) {
	viper.Set("expiry", expiry)
}

func GetPipelineClientID() string {
	return os.Getenv("clientid")
}

func GetPipelineClientSecret() string {
	return os.Getenv("clientsecret")
}

func PipelineLogin() error {
	uri, err := url.Parse(GetTokenUrl())
	if err != nil {
		return err
	}

	query := &url.Values{}
	query.Add("grant_type", "client_credentials")
	uri.RawQuery = query.Encode()

	data := &url.Values{}
	data.Add("client_id", GetPipelineClientID())
	data.Add("client_secret", GetPipelineClientSecret())

	ctx := context.TODO()
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

	SetPipelineToken(tResponse.AccessToken)
	SetPipelineTokenExpiry(now.Add(time.Second * time.Duration(tResponse.ExpiresIn)))

	return nil
}
