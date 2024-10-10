// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	v3 "github.com/sailpoint-oss/golang-sdk/v2/api_v3"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdk"
	"github.com/sailpoint-oss/sailpoint-cli/internal/search"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newPreviewCommand() *cobra.Command {
	var showLongCommand bool
	var resultOnly bool
	var filepath string
	var profile string
	var identity string
	var identityProfile *v2024.IdentityProfile
	var identityAttributeConfig *v2024.IdentityAttributeConfig
	var identityPreview *v2024.IdentityPreviewResponse
	cmd := &cobra.Command{
		Use:     "preview",
		Short:   "Preview a transform result in Identity Security Cloud",
		Long:    "\nPreview a transform result in Identity Security Cloud\n\n",
		Example: "sail transform preview | sail transform pre",
		Aliases: []string{"pre"},
		Args:    cobra.OnlyValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var transform v2024.Transform
			var decoder *json.Decoder

			if profile == "" && identity == "" {
				showLongCommand = true
			}

			if filepath != "" {
				file, err := os.Open(filepath)
				if err != nil {
					return err
				}
				defer file.Close()
				decoder = json.NewDecoder(bufio.NewReader(file))
			} else {
				log.Error("You must provide a file to preview")
				return nil
			}

			if err := decoder.Decode(&transform); err != nil {
				return err
			}

			log.Debug("Filepath", "path", filepath)

			log.Debug("Transform", "transform", transform)

			transform.SetName(transform.GetName() + "-preview")

			if transform.GetName() == "" {
				return fmt.Errorf("the transform must have a name")
			}

			apiClient, err := config.InitAPIClient(true)

			if err != nil {
				return err
			}

			transformObj, resp, err := apiClient.V2024.TransformsAPI.CreateTransform(context.TODO()).Transform(transform).Execute()

			defer cleanupPreviewObjects(apiClient, transformObj.GetId())

			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			var attributeType = "string"
			identityAttribute, resp, err := apiClient.Beta.IdentityAttributesAPI.CreateIdentityAttribute(context.TODO()).IdentityAttribute(beta.IdentityAttribute{Name: "sailpointCLIPreview", Type: *beta.NewNullableString(&attributeType)}).Execute()

			defer cleanupIdentityAttribute(apiClient, identityAttribute.GetName())

			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			if profile == "" {
				identity_profiles, resp, err := sailpoint.PaginateWithDefaults[v3.IdentityProfile](apiClient.V3.IdentityProfilesAPI.ListIdentityProfiles(context.TODO()))
				if err != nil {
					return sdk.HandleSDKError(resp, err)
				}

				profile, err = SelectProfile(identity_profiles)
				if err != nil {
					return err
				}
			}

			identityProfile, resp, err = apiClient.V2024.IdentityProfilesAPI.GetIdentityProfile(context.TODO(), profile).Execute()

			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			identityAttributeConfig = identityProfile.IdentityAttributeConfig

			if identity == "" {
				searchObj, err := search.BuildSearch(fmt.Sprintf("identityProfile.id:%s", profile), []string{"name"}, []string{"identities"})
				if err != nil {
					return err
				}

				formattedResponse, err := search.PerformSearch(*apiClient, searchObj)
				if err != nil {
					return err
				}

				identity, err = SelectIdentity(formattedResponse.Identities)
				if err != nil {
					return err
				}
			}

			var transformType = "reference"
			identityAttributeConfig.AttributeTransforms = append(identityAttributeConfig.AttributeTransforms, v2024.IdentityAttributeTransform{
				IdentityAttributeName: &identityAttribute.Name,
				TransformDefinition: &v2024.TransformDefinition{
					Type: &transformType,
					Attributes: map[string]interface{}{
						"id": transformObj.GetName(),
					},
				},
			})

			var enabled = true
			var request = v2024.IdentityPreviewRequest{
				IdentityId: &identity,
				IdentityAttributeConfig: &v2024.IdentityAttributeConfig{
					Enabled:             &enabled,
					AttributeTransforms: identityAttributeConfig.AttributeTransforms,
				},
			}

			identityPreview, resp, err = apiClient.V2024.IdentityProfilesAPI.ShowIdentityPreview(context.TODO()).IdentityPreviewRequest(request).Execute()

			if err != nil {
				//fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
			}

			var entries [][]string

			for _, v := range identityPreview.PreviewAttributes {
				if v.GetName() == "sailpointCLIPreview" {
					if v.GetErrorMessages() != nil {
						errorMap, err := v.GetErrorMessages()[0].ToMap()
						if err != nil {
							return err
						}
						log.Error("An error occurred while previewing the transform")
						print(util.RenderMarkdown("```json\n" + util.PrettyPrint(errorMap) + "\n```"))
					} else {
						if !resultOnly {
							log.Info("", "transform result", v.GetValue())
						} else {
							fmt.Println(v.GetValue())
						}
					}
				} else {
					if v.GetErrorMessages() != nil {
						entries = append(entries, []string{*v.Name, *v.GetErrorMessages()[0].Text})
					} else {
						entries = append(entries, []string{*v.Name, v.GetValue()})
					}

				}
			}

			if !resultOnly {
				output.WriteTable(cmd.OutOrStdout(), []string{"Attribute", "Value"}, entries, "Attribute")
			}

			if showLongCommand {
				fmt.Printf("Use the following command to preview the transform with this identity directly.\n\n")
				fmt.Printf("sail transform preview --profile %s --identity %s --file %s\n", profile, identity, filepath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filepath, "file", "f", "", "The path to the transform you wish to preview")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "The identity profile of the transform you wish to preview")
	cmd.Flags().StringVarP(&identity, "identity", "i", "", "The identity you wish to preview the transform with")
	cmd.Flags().BoolVarP(&resultOnly, "result-only", "r", false, "Only show the result of the transform")

	return cmd
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

func cleanupPreviewObjects(apiClient *sailpoint.APIClient, transformId string) error {
	log.Debug("Cleaning up preview objects")

	resp, err := apiClient.V2024.TransformsAPI.DeleteTransform(context.TODO(), transformId).Execute()
	if err != nil {
		return sdk.HandleSDKError(resp, err)
	}

	return nil
}

func cleanupIdentityAttribute(apiClient *sailpoint.APIClient, attributeName string) error {
	log.Debug("Cleaning up identity attribute object")

	resp, err := apiClient.Beta.IdentityAttributesAPI.DeleteIdentityAttribute(context.TODO(), attributeName).Execute()

	if err != nil {
		return sdk.HandleSDKError(resp, err)
	}

	return nil
}
