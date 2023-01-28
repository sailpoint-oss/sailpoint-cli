// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	sailpoint "github.com/sailpoint-oss/golang-sdk/sdk-output"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/root"
	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	c         client.Client
	apiClient sailpoint.APIClient
	rootCmd   *cobra.Command
)

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(filepath.Join(home, ".sailpoint"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("sail")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// IGNORE they may be using env vars
		} else {
			// Config file was found but another error was produced
			cobra.CheckErr(err)
		}
	}
}

func init() {
	initConfig()

	var config types.OrgConfig

	err := viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("unable to decode config: %s ", err))
	}

	auth.EnsureAccessToken(config, context.TODO())
	c = client.NewSpClient(config)

	var DevNull types.DevNull

	configuration := sailpoint.NewConfiguration(sailpoint.ClientConfiguration{Token: util.GetAuthToken(), BaseURL: util.GetBaseUrl()})
	apiClient := sailpoint.NewAPIClient(configuration)
	apiClient.V3.GetConfig().HTTPClient.Logger = DevNull
	apiClient.Beta.GetConfig().HTTPClient.Logger = DevNull

	rootCmd = root.NewRootCmd(c, apiClient)

}

// main the entry point for commands. Note that we do not need to do cobra.CheckErr(err)
// here. When a command returns error, cobra already logs it. Adding CheckErr here will
// cause error messages to be logged twice. We do need to exit with error code if something
// goes wrong. This will exit the cli container during pipeline build and fail that stage.
func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
