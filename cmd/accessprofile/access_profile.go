package accessprofile

import (
	"context"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdkcmd"
	"github.com/spf13/cobra"
)

func NewAccessProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "access-profile",
		Aliases: []string{"accessprofile"},
		Short:   "Inspect access profiles",
		Long:    "\nInspect Identity Security Cloud access profiles and their entitlements.\n\n",
		Example: "  sail access-profile list\n  sail access-profile get <access-profile-id>",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(newListCommand(), newGetCommand(), newCreateCommand(), newPatchCommand(), newDeleteCommand(), newEntitlementsCommand())
	return cmd
}

func newListCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List access profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.AccessProfilesAPI.ListAccessProfiles(context.TODO()).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			if opts.Filters != "" {
				req = req.Filters(opts.Filters)
			}
			if opts.Sorters != "" {
				req = req.Sorters(opts.Sorters)
			}
			profiles, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			rows := make([][]string, 0, len(profiles))
			for _, item := range profiles {
				rows = append(rows, []string{item.GetName(), item.GetId(), item.GetDescription()})
			}
			return sdkcmd.WriteTable(cmd, []string{"Name", "ID", "Description"}, rows, "Name", profiles)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <access-profile-id>",
		Short: "Get an access profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			profile, resp, err := apiClient.V2024.AccessProfilesAPI.GetAccessProfile(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, profile)
		},
	}
}

func newCreateCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create --file access-profile.json",
		Short: "Create an access profile from a JSON payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.AccessProfile](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			profile, resp, err := apiClient.V2024.AccessProfilesAPI.CreateAccessProfile(context.TODO()).
				AccessProfile(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, profile)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newPatchCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:     "patch <access-profile-id> --file patch.json",
		Aliases: []string{"update"},
		Short:   "Patch an access profile from a JSON Patch payload",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[[]v2024.JsonPatchOperation](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			profile, resp, err := apiClient.V2024.AccessProfilesAPI.PatchAccessProfile(context.TODO(), args[0]).
				JsonPatchOperation(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, profile)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON Patch payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newDeleteCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <access-profile-id>",
		Short: "Delete an access profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return clierror.Usage("access-profile delete requires --force")
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			resp, err := apiClient.V2024.AccessProfilesAPI.DeleteAccessProfile(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, map[string]string{"status": "deleted", "accessProfileId": args[0]})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm access profile deletion")
	return cmd
}

func newEntitlementsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "entitlements <access-profile-id>",
		Short: "List entitlements for an access profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			entitlements, resp, err := apiClient.V2024.AccessProfilesAPI.GetAccessProfileEntitlements(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, entitlements)
		},
	}
}
