// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/root"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	c       client.Client
	rootCmd *cobra.Command
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
		panic(fmt.Errorf("Unable to decode Config: %s \n", err))
	}

	c = client.NewSpClient(types.OrgConfig{
		AuthType: config.AuthType,
		Debug:    config.Debug,
		Pat:      config.Pat,
		OAuth:    config.OAuth,
	})
	rootCmd = root.NewRootCmd(c)

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
