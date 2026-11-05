// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/cmd/root"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command
var stateless bool

func init() {

	rootCmd = root.NewRootCommand()
	command, _, err := rootCmd.Find(os.Args[1:])
	stateless = err == nil && command.Annotations["sailpoint.stateless"] == "true"
	// Version reporting must not read or rewrite credential configuration.
	stateless = stateless || (len(os.Args) == 2 && os.Args[1] == "--version")
	if !stateless {
		cobra.CheckErr(config.InitConfig())
	}

}

// main the entry point for commands. Note that we do not need to do cobra.CheckErr(err)
// here. When a command returns error, cobra already logs it. Adding CheckErr here will
// cause error messages to be logged twice. We do need to exit with error code if something
// goes wrong. This will exit the cli container during pipeline build and fail that stage.
func main() {

	err := rootCmd.Execute()

	if !stateless {
		if saveErr := config.SaveConfig(); saveErr != nil {
			log.Warn("Issue saving config file", "error", saveErr)
		}
	}

	// When error occurs, we need to make sure we exit the program with an error code. We
	// don't need to log it here because the sub commands already log it to the console.
	if err != nil {
		os.Exit(1)
	}
}
