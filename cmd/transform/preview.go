// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v3"
	"github.com/sailpoint-oss/golang-sdk/v3/identity_attributes"
	"github.com/sailpoint-oss/golang-sdk/v3/identity_profiles"
	"github.com/sailpoint-oss/golang-sdk/v3/transforms"

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
	var identityProfile *identity_profiles.Identityprofile
	var identityAttributeConfig *identity_profiles.Identityattributeconfig
	var identityPreview *identity_profiles.Identitypreviewresponse
	cmd := &cobra.Command{
		Use:     "preview",
		Short:   "Preview a transform result in Identity Security Cloud",
		Long:    "\nPreview a transform result in Identity Security Cloud\n\n",
		Example: "sail transform preview | sail transform pre",
		Aliases: []string{"pre"},
		Args:    cobra.OnlyValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var transform transforms.Transform
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

			transformObj, resp, err := apiClient.TransformsAPI.CreateTransformV1(context.TODO()).Transform(transform).Execute()

			defer cleanupPreviewObjects(apiClient, transformObj.GetId())

			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			var attributeType = "string"
			identityAttribute, resp, err := apiClient.IdentityAttributesAPI.CreateIdentityAttributeV1(context.TODO()).Identityattribute2(identity_attributes.Identityattribute2{Name: "sailpointCLIPreview", Type: *identity_attributes.NewNullableString(&attributeType)}).Execute()

			defer cleanupIdentityAttribute(apiClient, identityAttribute.GetName())

			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			if profile == "" {
				profileList, resp, err := sailpoint.PaginateWithDefaults[identity_profiles.Identityprofile](apiClient.IdentityProfilesAPI.ListIdentityProfilesV1(context.TODO()))
				if err != nil {
					return sdk.HandleSDKError(resp, err)
				}

				profile, err = SelectProfile(profileList)
				if err != nil {
					return err
				}
			}

			identityProfile, resp, err = apiClient.IdentityProfilesAPI.GetIdentityProfileV1(context.TODO(), profile).Execute()

			if err != nil {
				return sdk.HandleSDKError(resp, err)
			}

			identityAttributeConfig = identityProfile.IdentityAttributeConfig

			if identity == "" {
				searchObj, err := search.BuildSearch(fmt.Sprintf("identityProfile.id:%s", profile), []string{"name"}, []string{"identities"})
				if err != nil {
					return err
				}

				searchObj.QueryResultFilter.Includes = []string{"id", "displayName", "email"}

				formattedResponse, err := search.PerformSearchWithLimit(*apiClient, searchObj, 250)
				if err != nil {
					return err
				}

				identity, err = SelectIdentity(formattedResponse.Identities)
				if err != nil {
					return err
				}
			}

			var transformType = "reference"
			identityAttributeConfig.AttributeTransforms = append(identityAttributeConfig.AttributeTransforms, identity_profiles.Identityattributetransform{
				IdentityAttributeName: &identityAttribute.Name,
				TransformDefinition: &identity_profiles.Transformdefinition{
					Type: &transformType,
					Attributes: map[string]interface{}{
						"id": transformObj.GetName(),
					},
				},
			})

			var enabled = true
			var request = identity_profiles.Identitypreviewrequest{
				IdentityId: &identity,
				IdentityAttributeConfig: &identity_profiles.Identityattributeconfig{
					Enabled:             &enabled,
					AttributeTransforms: identityAttributeConfig.AttributeTransforms,
				},
			}

			identityPreview, resp, err = apiClient.IdentityProfilesAPI.GenerateIdentityPreviewV1(context.TODO()).Identitypreviewrequest(request).Execute()

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

func SelectProfile(profiles []identity_profiles.Identityprofile) (string, error) {
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

	resp, err := apiClient.TransformsAPI.DeleteTransformV1(context.TODO(), transformId).Execute()
	if err != nil {
		return sdk.HandleSDKError(resp, err)
	}

	return nil
}

func cleanupIdentityAttribute(apiClient *sailpoint.APIClient, attributeName string) error {
	log.Debug("Cleaning up identity attribute object")

	resp, err := apiClient.IdentityAttributesAPI.DeleteIdentityAttributeV1(context.TODO(), attributeName).Execute()

	if err != nil {
		return sdk.HandleSDKError(resp, err)
	}

	return nil
}
