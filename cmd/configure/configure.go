// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package configure

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/sailpoint-oss/sailpoint-cli/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"
)

const (
	baseURLTemplate  = "https://%s"
	tokenURLTemplate = "https://%s/oauth/token"
	authURLTemplate  = "https://%s/oauth/authorize"
	configFolder     = ".sailpoint"
	configYamlFile   = "config.yaml"
)

var (
	baseURL  (string)
	tokenURL (string)
	authURL  (string)
)

func PromptAuth() (string, error) {
	items := []types.Choice{
		{Title: "PAT", Description: "Person Access Token - Single User"},
		{Title: "OAuth", Description: "OAuth2.0 Authentication - Sign in via the Website"},
	}

	choice, err := tui.PromptList(items, "Choose an authentication method to configure")
	if err != nil {
		return "", err
	}

	return strings.ToLower(choice.Title), nil
}

func NewConfigureCmd(client client.Client) *cobra.Command {
	var debug bool
	cmd := &cobra.Command{
		Use:     "configure",
		Short:   "configure authentication for the cli",
		Long:    "\nConfigure Authentication for the CLI\nSupported Methods: (PAT, OAuth)\n\nPrerequisites:\n\nPAT:\n	Tenant\n	Client ID\n	Client Secret\n\nOAuth:\n	Tenant\n	Client ID\n	Client Secret - Optional Depending on configuration\n	Callback Port (ex. http://localhost:{3000}/callback)\n	Callback Path (ex. http://localhost:3000{/callback})",
		Aliases: []string{"conf"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			var AuthType string
			var err error

			if len(args) > 0 {
				AuthType = args[0]
			} else {
				AuthType, err = PromptAuth()
				if err != nil {
					return err
				}
			}

			config, err := getConfigureParamsFromStdin(AuthType, debug)
			if err != nil {
				return err
			}

			err = updateConfigFile(config)
			if err != nil {
				return err
			}

			switch strings.ToLower(AuthType) {
			case "pat":
				err = config.PATLogin()
				if err != nil {
					return err
				}
			case "oauth":
				err := config.OAuthLogin()
				if err != nil {
					return err
				}
			default:
				return errors.New("invalid authtype")
			}

			return nil

		},
	}
	cmd.Flags().BoolVarP(&debug, "Debug", "d", false, "Specifies if the debug flag should be set")

	return cmd
}

func updateConfigFile(conf types.CLIConfig) error {
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

	viper.Set(fmt.Sprintf("environments.%s", conf.ActiveEnvironment), conf.Environments[conf.ActiveEnvironment])

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

func setIDNUrls(tenant string) {
	var tokens = strings.Split(tenant, ".")
	tokens = append(tokens[:1+1], tokens[1:]...)
	tokens[1] = "api"
	var api_base = strings.Join(tokens, ".")
	baseURL = fmt.Sprintf(baseURLTemplate, api_base)
	tokenURL = fmt.Sprintf(tokenURLTemplate, api_base)
	authURL = fmt.Sprintf(authURLTemplate, tenant)
}

func getConfigureParamsFromStdin(AuthType string, debug bool) (types.CLIConfig, error) {
	var config types.CLIConfig

	switch strings.ToLower(AuthType) {
	case "pat":
		var Pat types.PatConfig
		paramsNames := []string{
			"Tenant (ex. tenant.identitynow.com): ",
			"Personal Access Token Client ID: ",
			"Personal Access Token Client Secret: ",
		}

		scanner := bufio.NewScanner(os.Stdin)
		for _, pm := range paramsNames {
			fmt.Print(pm)
			scanner.Scan()
			value := scanner.Text()

			if value == "" {
				return config, fmt.Errorf("%s parameter is empty", pm[:len(pm)-2])
			}

			switch pm {
			case paramsNames[0]:
				setIDNUrls(value)
				Pat.Tenant = value
				Pat.BaseUrl = baseURL
				Pat.TokenUrl = tokenURL
			case paramsNames[1]:
				Pat.ClientID = value
			case paramsNames[2]:
				Pat.ClientSecret = value
			}
		}
		config.AuthType = AuthType
		tempEnv := config.Environments[config.ActiveEnvironment]
		tempEnv.Pat = Pat
		config.Environments[config.ActiveEnvironment] = tempEnv

		return config, nil
	case "oauth":
		var OAuth types.OAuthConfig
		paramsNames := []string{
			"Tenant (ex. tenant.identitynow.com): ",
			"OAuth Client ID: ",
			"OAuth Client Secret: ",
			"OAuth Redirect Port (ex. http://localhost:{3000}/callback): ",
			"OAuth Redirect Path (ex. http://localhost:3000{/callback}): ",
		}

		scanner := bufio.NewScanner(os.Stdin)
		for _, pm := range paramsNames {
			fmt.Print(pm)
			scanner.Scan()
			value := scanner.Text()

			if value == "" && pm != paramsNames[2] {
				return config, fmt.Errorf("%s parameter is empty", pm[:len(pm)-2])
			}

			switch pm {
			case paramsNames[0]:
				setIDNUrls(value)
				OAuth.Tenant = value
				OAuth.BaseUrl = baseURL
				OAuth.TokenUrl = tokenURL
				OAuth.AuthUrl = authURL
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
		config.AuthType = AuthType
		tempEnv := config.Environments[config.ActiveEnvironment]
		tempEnv.OAuth = OAuth
		config.Environments[config.ActiveEnvironment] = tempEnv

		return config, nil
	default:
		return config, errors.New("invalid auth type provided")
	}
}
