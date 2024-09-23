// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"fmt"

	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	v3 "github.com/sailpoint-oss/golang-sdk/v2/api_v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/search"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/spf13/cobra"
)

func newPreviewCommand() *cobra.Command {
	var profile string
	var identity string
	return &cobra.Command{
		Use:     "preview",
		Short:   "Preview a transform result in Identity Security Cloud",
		Long:    "\nPreview a transform result in Identity Security Cloud\n\n",
		Example: "sail transform preview | sail transform pre",
		Aliases: []string{"pre"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			identity_profiles, resp, err := sailpoint.PaginateWithDefaults[v3.IdentityProfile](apiClient.V3.IdentityProfilesAPI.ListIdentityProfiles(context.TODO()))
			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			profile, err = SelectProfile(identity_profiles)
			if err != nil {
				return err
			}

			searchObj, err := search.BuildSearch(fmt.Sprintf("identityProfile.id:%s", profile), []string{"name"}, []string{"identities"})
			if err != nil {
				return err
			}

			print(*searchObj.Query.Query)

			formattedResponse, err := search.PerformSearch(*apiClient, searchObj)
			if err != nil {
				return err
			}

			identity, err = SelectIdentity(formattedResponse.Identities)
			if err != nil {
				return err
			}

			fmt.Print(identity)

			// var entries [][]string

			// for _, v := range identity_profiles {
			// 	entries = append(entries, []string{v.Name, *v.Id})
			// }

			// output.WriteTable(cmd.OutOrStdout(), []string{"Name", "ID"}, entries)

			return nil
		},
	}
}

func SelectProfile[T v3.IdentityProfile](profiles []v3.IdentityProfile) (string, error) {
	var prompts []tui.Choice
	for i := 0; i < len(profiles); i++ {
		temp := profiles[i]

		prompts = append(prompts, tui.Choice{Title: temp.GetName(), Description: temp.GetDescription(), Id: temp.GetId()})
	}

	intermediate, err := tui.PromptList(prompts, "Select an Identity Profile to preview the transform")
	if err != nil {
		return "", err
	}
	return intermediate.Id, nil

}

func SelectIdentity[T search.Identity](identities []search.Identity) (string, error) {
	var prompts []tui.Choice
	for i := 0; i < len(identities); i++ {
		temp := identities[i]

		prompts = append(prompts, tui.Choice{Title: temp.DisplayName, Description: temp.Email, Id: temp.ID})
	}

	intermediate, err := tui.PromptList(prompts, "Select an Identity to preview the transform")
	if err != nil {
		return "", err
	}
	return intermediate.Id, nil

}
