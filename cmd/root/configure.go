// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package root

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	"github.com/sailpoint-oss/sailpoint-cli/auth"
	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/sailpoint-oss/sailpoint-cli/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"
)

const (
	baseURLTemplate  = "https://%s.api.identitynow.com"
	tokenURLTemplate = "https://%s.api.identitynow.com/oauth/token"
	authUrlTemplate  = "https://%s.identitynow.com/oauth/authorize"
	configFolder     = ".sailpoint"
	configYamlFile   = "config.yaml"
)

func newConfigureCmd(client client.Client) *cobra.Command {
	var AuthType string
	var debug bool
	cmd := &cobra.Command{
		Use:     "configure",
		Short:   "Configure CLI",
		Aliases: []string{"conf"},
		RunE: func(cmd *cobra.Command, args []string) error {

			fmt.Println(AuthType)

			config, err := getConfigureParamsFromStdin(AuthType, debug)
			if err != nil {
				return err
			}

			fmt.Println("Finished getting config")

			err = updateConfigFile(config)
			if err != nil {
				return err
			}

			fmt.Println("Finished writing config")

			switch AuthType {
			case "PAT":
				_, err = auth.PATLogin(config, cmd.Context())
				if err != nil {
					return err
				}
			case "OAuth":
				_, err := auth.OAuthLogin(config)
				if err != nil {
					return err
				}
			default:
				return errors.New("invalid authtype")
			}

			return nil

		},
	}
	cmd.Flags().StringVarP(&AuthType, "AuthType", "t", "", "Specifies the Authentication method that should be configured (PAT, OAuth)")
	cmd.Flags().BoolVarP(&debug, "Debug", "d", false, "Specifies if the debug flag should be set")

	return cmd
}

func updateConfigFile(conf types.OrgConfig) error {
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

	viper.Set("authtype", conf.AuthType)
	viper.Set("debug", conf.Debug)

	switch conf.AuthType {
	case "PAT":
		viper.Set("pat", conf.Pat)
	case "OAuth":
		viper.Set("oauth", conf.OAuth)
	}

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

func getConfigureParamsFromStdin(AuthType string, debug bool) (types.OrgConfig, error) {
	var conf types.OrgConfig

	switch AuthType {
	case "PAT":
		paramsNames := []string{
			"Tenant (ex. {tenant}.identitynow.com): ",
			"Personal Access Token Client ID: ",
			"Personal Access Token Client Secret: ",
		}

		scanner := bufio.NewScanner(os.Stdin)
		for _, pm := range paramsNames {
			fmt.Print(pm)
			scanner.Scan()
			value := scanner.Text()

			if value == "" {
				return conf, fmt.Errorf("%s parameter is empty", pm[:len(pm)-2])
			}

			switch pm {
			case paramsNames[0]:
				conf.Pat.Tenant = value
				conf.Pat.BaseUrl = fmt.Sprintf(baseURLTemplate, value)
				conf.Pat.TokenUrl = fmt.Sprintf(tokenURLTemplate, value)
			case paramsNames[1]:
				conf.Pat.ClientID = value
			case paramsNames[2]:
				conf.Pat.ClientSecret = value
			}
		}
		conf.AuthType = AuthType

		fmt.Println("got here")

		return conf, nil
	case "OAuth":
		paramsNames := []string{
			"Tenant (ex. {tenant}.identitynow.com): ",
			"OAuth Client ID: ",
			"OAuth Client Secret: ",
			"OAuth Redirect Port (ex. http://localhost:{3000}/callback): ",
			"OAuth Redirect Path (ex. http://localhost:3000{/callback}): ",
		}

		scanner := bufio.NewScanner(os.Stdin)
		var OAuth types.OAuthConfig
		for _, pm := range paramsNames {
			fmt.Print(pm)
			scanner.Scan()
			value := scanner.Text()

			if value == "" && pm != paramsNames[2] {
				return conf, fmt.Errorf("%s parameter is empty", pm[:len(pm)-2])
			}

			switch pm {
			case paramsNames[0]:
				OAuth.Tenant = value
				OAuth.BaseUrl = fmt.Sprintf(baseURLTemplate, value)
				OAuth.TokenUrl = fmt.Sprintf(tokenURLTemplate, value)
				OAuth.AuthUrl = fmt.Sprintf(authUrlTemplate, value)
			case paramsNames[1]:
				OAuth.ClientID = value
			case paramsNames[2]:
				OAuth.ClientSecret = value
			case paramsNames[3]:
				OAuth.Redirect.Port, _ = strconv.Atoi(value)
			case paramsNames[4]:
				OAuth.Redirect.Path = value
			}
		}
		conf.OAuth = OAuth
		conf.AuthType = AuthType

		return conf, nil
	}
	return conf, errors.New("invalid auth type provided")
}
