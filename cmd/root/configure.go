// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package root

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"
)

const (
	baseURLTemplate  = "https://%s.api.identitynow.com"
	tokenURLTemplate = "%s/oauth/token"
	configFolder     = ".sailpoint"
	configYamlFile   = "config.yaml"
)

type OrgConfig struct {
	BaseUrl      string `mapstructure:"baseURL"`
	TokenUrl     string `mapstructure:"tokenURL"`
	ClientSecret string `mapstructure:"clientSecret"`
	ClientID     string `mapstructure:"clientID"`
	Debug        bool   `mapstructure:"debug"`
}

func newConfigureCmd(client client.Client) *cobra.Command {
	conn := &cobra.Command{
		Use:     "configure",
		Short:   "Configure CLI",
		Aliases: []string{"conf"},
		RunE: func(cmd *cobra.Command, args []string) error {

			config, err := getConfigureParamsFromStdin()
			if err != nil {
				return err
			}

			err = updateConfigFile(config)
			if err != nil {
				return err
			}

			err = client.VerifyToken(context.Background(), config.TokenUrl, config.ClientID, config.ClientSecret)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return conn
}

func updateConfigFile(conf *OrgConfig) error {
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

	viper.Set("baseUrl", conf.BaseUrl)
	viper.Set("tokenUrl", conf.TokenUrl)
	viper.Set("clientSecret", conf.ClientSecret)
	viper.Set("clientID", conf.ClientID)
	viper.Set("debug", false)

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

func getConfigureParamsFromStdin() (*OrgConfig, error) {
	conf := &OrgConfig{}

	paramsNames := []string{
		"Base URL (ex. https://{org}.api.identitynow.com): ",
		"Personal Access Token Client ID: ",
		"Personal Access Token Client Secret: ",
	}

	scanner := bufio.NewScanner(os.Stdin)
	for _, pm := range paramsNames {
		fmt.Print(pm)
		scanner.Scan()
		value := scanner.Text()

		if value == "" {
			return nil, fmt.Errorf("%s parameter is empty", pm[:len(pm)-2])
		}

		switch pm {
		case paramsNames[0]:
			conf.BaseUrl = value
			conf.TokenUrl = fmt.Sprintf(tokenURLTemplate, conf.BaseUrl)
		case paramsNames[1]:
			conf.ClientID = value
		case paramsNames[2]:
			conf.ClientSecret = value
		}
	}

	return conf, nil
}
