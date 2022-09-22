// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/sailpoint-oss/sp-cli/client"
	"github.com/sailpoint-oss/sp-cli/util"
	"github.com/spf13/cobra"
)

func newPreviewCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "preview -i <identity-profile-id> -a <attribute-name> -n <transform-name>",
		Short:   "Preview transform",
		Long:    "Preview the final output of a transform",
		Example: "sp transforms preview -i 12a199b967b64ffe992ef4ecfd076728 -a lastname -n ToLower",
		Aliases: []string{"p"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			idProfile := cmd.Flags().Lookup("identity-profile").Value.String()
			if idProfile == "" {
				return fmt.Errorf("identity-profile must be specified")
			}

			attribute := cmd.Flags().Lookup("attribute").Value.String()
			if attribute == "" {
				return fmt.Errorf("attribute must be specified")
			}

			name := cmd.Flags().Lookup("name").Value.String()
			if name == "" {
				return fmt.Errorf("name must be specified")
			}

			// Get the identity profile so we can obtain the authoritative source and
			// original transform for the attribute, which will contain the account attribute
			// name and source name that will be used in the preview body.
			endpoint := cmd.Flags().Lookup("identity-profile-endpoint").Value.String()
			resp, err := client.Get(cmd.Context(), util.ResourceUrl(endpoint, idProfile))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("Get identity profile failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var profile identityProfile
			err = json.Unmarshal(raw, &profile)
			if err != nil {
				return err
			}

			// Get a list of users in the source specified by the identity profile.
			// These users will be used to preview the transform.
			endpoint = cmd.Flags().Lookup("user-endpoint").Value.String()
			uri, err := url.Parse(endpoint)
			if err != nil {
				return err
			}

			query := &url.Values{}
			query.Add("filters", "[{\"property\":\"links.application.id\",\"operation\":\"EQ\",\"value\":\""+profile.AuthoritativeSource.Id+"\"}]")
			uri.RawQuery = query.Encode()

			resp, err = client.Get(cmd.Context(), uri.String())
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("Get users failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err = io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var user []user
			err = json.Unmarshal(raw, &user)
			if err != nil {
				return err
			}

			// enc := json.NewEncoder(os.Stdout)
			// enc.SetIndent("", "    ")
			// if err := enc.Encode(previewBody); err != nil {
			// 	log.Fatal(err)
			// }

			// Form the request body that will be sent to the preview endpoint
			var accountAttName string
			var sourceName string
			for _, t := range profile.IdentityAttributeConfig.AttributeTransforms {
				if t.IdentityAttributeName == attribute {
					transType := t.TransformDefinition.Type
					if transType == "accountAttribute" {
						def := makeAttributesOfAccount(t.TransformDefinition.Attributes)
						accountAttName = def.AttributeName
						sourceName = def.SourceName
					} else if transType == "reference" {
						def := makeReference(t.TransformDefinition.Attributes)
						accountAttName = def.Input.Attributes.AttributeName
						sourceName = def.Input.Attributes.SourceName
					} else {
						log.Fatal("Unknown transform definition encountered when parsing identity profile: " + transType)
						return nil
					}
				}
			}

			previewBody := makePreviewBody(attribute, name, accountAttName, sourceName)

			raw, err = json.Marshal(previewBody)
			if err != nil {
				return err
			}

			// Call the preview endpoint to get the raw and transformed attribute values
			endpoint = cmd.Flags().Lookup("preview-endpoint").Value.String()
			resp, err = client.Post(cmd.Context(), util.ResourceUrl(endpoint, user[0].Id), "application/json", bytes.NewReader(raw))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("preview transform failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err = io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var response previewResponse
			err = json.Unmarshal(raw, &response)
			if err != nil {
				return err
			}

			for _, x := range response.PreviewAttributes {
				if x.Name == attribute {
					fmt.Printf("Original value: %s\nTransformed value: %s\n", x.PreviousValue, x.Value)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("identity-profile", "i", "", "The GUID of an identity profile (required)")
	cmd.Flags().StringP("attribute", "a", "", "Attribute name (required)")
	cmd.Flags().StringP("name", "n", "", "Transform name (required)")
	// cmd.Flags().StringP("file", "f", "", "The path to the transform file (required)")

	cmd.MarkFlagRequired("identity-profile")
	cmd.MarkFlagRequired("attribute")
	cmd.MarkFlagRequired("name")
	// cmd.MarkFlagRequired("file")

	return cmd
}
