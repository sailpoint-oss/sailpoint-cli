// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package main

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/root"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func init() {

	cobra.CheckErr(config.InitConfig())
	rootCmd = root.NewRootCommand()

}

// main the entry point for commands. Note that we do not need to do cobra.CheckErr(err)
// here. When a command returns error, cobra already logs it. Adding CheckErr here will
// cause error messages to be logged twice. We do need to exit with error code if something
// goes wrong. This will exit the cli container during pipeline build and fail that stage.
func main() {

	err := rootCmd.Execute()

	if saveErr := config.SaveConfig(); saveErr != nil {
		log.Warn("Issue saving config file", "error", saveErr)
	}

	if err != nil {
		log.Fatal(err)
	}
}
