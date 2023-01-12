package util

import (
	"strings"

	"github.com/spf13/viper"
)

func GetAuthType() string {
	return strings.ToLower(viper.GetString("authtype"))
}

func GetBasePath() string {
	switch GetAuthType() {
	case "oauth":
		return viper.GetString("oauth.baseurl")
	case "pat":
		return viper.GetString("pat.baseurl")
	}
	return ""
}
