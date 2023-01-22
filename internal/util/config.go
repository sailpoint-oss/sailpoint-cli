package util

import (
	"strings"

	"github.com/spf13/viper"
)

func GetAuthType() string {
	return strings.ToLower(viper.GetString("authtype"))
}

func GetBaseUrl() string {
	switch GetAuthType() {
	case "oauth":
		return viper.GetString("oauth.baseurl")
	case "pat":
		return viper.GetString("pat.baseurl")
	}
	return ""
}

func GetAuthToken() string {
	switch GetAuthType() {
	case "oauth":
		return viper.GetString("oauth.token.accesstoken")
	case "pat":
		return viper.GetString("pat.token.accesstoken")
	}
	return ""
}
