// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	transmodel "github.com/sailpoint-oss/sailpoint-cli/cmd/transform/model"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var implicitInput bool

func newPreviewCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "preview",
		Short:   "Preview transform",
		Long:    "Preview the final output of a transform.",
		Example: "sail transform p -i 12a199b967b64ffe992ef4ecfd076728 -a lastname -f /path/to/transform.json\nsail transform p -i 12a199b967b64ffe992ef4ecfd076728 -a lastname -n ToLower --implicit",
		Aliases: []string{"p"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			idProfile := cmd.Flags().Lookup("identity-profile").Value.String()
			attribute := cmd.Flags().Lookup("attribute").Value.String()

			var transform map[string]interface{}

			if !implicitInput {
				filepath := cmd.Flags().Lookup("file").Value.String()
				if filepath != "" {
					file, err := os.Open(filepath)
					if err != nil {
						return err
					}
					defer file.Close()

					err = json.NewDecoder(file).Decode(&transform)
					if err != nil {
						return err
					}
				} else {
					err := json.NewDecoder(os.Stdin).Decode(&transform)
					if err != nil {
						return err
					}
				}
			}

			// Get the identity profile so we can obtain the authoritative source and
			// original transform for the attribute, which will contain the account attribute
			// name and source name that will be used in the preview body.
			endpoint := viper.GetString("baseurl") + identityProfileEndpoint
			resp, err := client.Get(cmd.Context(), util.ResourceUrl(endpoint, idProfile))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("get identity profile failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var profile transmodel.IdentityProfile
			err = json.Unmarshal(raw, &profile)
			if err != nil {
				return err
			}

			// Get a list of identities in the authoritative source attached to the identity profile.
			// These identities will be used to preview the transform.
			endpoint = viper.GetString("baseurl") + searchEndpoint
			data := transmodel.MakeSearchQuery(profile.AuthoritativeSource.Id)
			raw, err = json.Marshal(data)
			if err != nil {
				return err
			}
			resp, err = client.Post(cmd.Context(), util.ResourceUrl(endpoint), "application/json", bytes.NewReader(raw))
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("get identity failed. status: %s\nbody: %s", resp.Status, body)
			}

			raw, err = io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var identity []transmodel.Identity
			err = json.Unmarshal(raw, &identity)
			if err != nil {
				return err
			}

			var previewBodyRaw []byte
			// If using implicit input, then attempt to grab the implicit
			// input from the identity profile mapping.
			if implicitInput {
				var accountAttName string
				var sourceName string
				for _, t := range profile.IdentityAttributeConfig.AttributeTransforms {
					if t.IdentityAttributeName == attribute {
						transType := t.TransformDefinition.Type
						if transType == "accountAttribute" {
							def := transmodel.MakeAttributesOfAccount(t.TransformDefinition.Attributes)
							accountAttName = def.AttributeName
							sourceName = def.SourceName
						} else if transType == "reference" {
							def := transmodel.MakeReference(t.TransformDefinition.Attributes)
							accountAttName = def.Input.Attributes.AttributeName
							sourceName = def.Input.Attributes.SourceName
						} else {
							return fmt.Errorf("Unknown transform definition encountered when parsing identity profile: " + transType)
						}
					}
				}

				name := cmd.Flags().Lookup("name").Value.String()
				if name == "" {
					return fmt.Errorf("the transform name must be specified when previewing with implicit input")
				}

				previewBody := transmodel.MakePreviewBodyImplicit(attribute, name, accountAttName, sourceName)

				previewBodyRaw, err = json.Marshal(previewBody)
				if err != nil {
					return err
				}
			} else {
				previewBody := transmodel.MakePreviewBodyExplicit(attribute, transform)

				previewBodyRaw, err = json.Marshal(previewBody)
				if err != nil {
					return err
				}
			}

			// Call the preview endpoint to get the raw and transformed attribute values
			endpoint = cmd.Flags().Lookup("preview-endpoint").Value.String()
			resp, err = client.Post(cmd.Context(), util.ResourceUrl(endpoint, identity[0].Id), "application/json", bytes.NewReader(previewBodyRaw))
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

			var response transmodel.PreviewResponse
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
	cmd.Flags().StringP("name", "n", "", "Transform name.  Only needed if using implicit input.  The transform must be uploaded to IDN first.")
	cmd.Flags().BoolVar(&implicitInput, "implicit", false, "Use implicit input.  Default is explicit input defined by the transform.")
	cmd.Flags().StringP("file", "f", "", "The path to the transform file.  Only needed if using explicit input.")

	cmd.MarkFlagRequired("identity-profile")
	cmd.MarkFlagRequired("attribute")

	return cmd
}
