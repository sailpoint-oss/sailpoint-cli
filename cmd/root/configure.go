// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package root

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sailpoint-oss/sailpoint-cli/internal/auth"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	tuilist "github.com/sailpoint-oss/sailpoint-cli/internal/tui/list"
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
	items := []list.Item{
		tuilist.Item("PAT"),
		tuilist.Item("OAuth"),
	}

	const defaultWidth = 20

	l := list.New(items, tuilist.ItemDelegate{}, defaultWidth, tuilist.ListHeight)
	l.Title = "What authentication method do you want to use?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = tuilist.TitleStyle
	l.Styles.PaginationStyle = tuilist.PaginationStyle
	l.Styles.HelpStyle = tuilist.HelpStyle

	m := tuilist.Model{List: l}
	_, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}

	choice := m.Retrieve()

	return choice, nil
}

func newConfigureCmd(client client.Client) *cobra.Command {
	var debug bool
	cmd := &cobra.Command{
		Use:     "configure",
		Short:   "Configure Authentication for the CLI",
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
				_, err = auth.PATLogin(config, cmd.Context())
				if err != nil {
					return err
				}
			case "oauth":
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

	switch strings.ToLower(conf.AuthType) {
	case "pat":
		viper.Set("pat", conf.Pat)
	case "oauth":
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

func setIDNUrls(tenant string) {
	var tokens = strings.Split(tenant, ".")
	tokens = append(tokens[:1+1], tokens[1:]...)
	tokens[1] = "api"
	var api_base = strings.Join(tokens, ".")
	baseURL = fmt.Sprintf(baseURLTemplate, api_base)
	tokenURL = fmt.Sprintf(tokenURLTemplate, api_base)
	authURL = fmt.Sprintf(authURLTemplate, tenant)
}

func getConfigureParamsFromStdin(AuthType string, debug bool) (types.OrgConfig, error) {
	var conf types.OrgConfig

	switch strings.ToLower(AuthType) {
	case "pat":
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
				return conf, fmt.Errorf("%s parameter is empty", pm[:len(pm)-2])
			}

			switch pm {
			case paramsNames[0]:
				setIDNUrls(value)
				conf.Pat.Tenant = value
				conf.Pat.BaseUrl = baseURL
				conf.Pat.TokenUrl = tokenURL
			case paramsNames[1]:
				conf.Pat.ClientID = value
			case paramsNames[2]:
				conf.Pat.ClientSecret = value
			}
		}
		conf.AuthType = AuthType

		return conf, nil
	case "oauth":
		paramsNames := []string{
			"Tenant (ex. tenant.identitynow.com): ",
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
		conf.OAuth = OAuth
		conf.AuthType = AuthType

		return conf, nil
	}
	return conf, errors.New("invalid auth type provided")
}
