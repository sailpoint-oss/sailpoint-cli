package role

import (
	"context"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdkcmd"
	"github.com/spf13/cobra"
)

func NewRoleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "role",
		Short:   "Inspect roles",
		Long:    "\nInspect Identity Security Cloud roles, role entitlements, and role members.\n\n",
		Example: "  sail role list\n  sail role get <role-id>\n  sail role entitlements <role-id>",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(newListCommand(), newGetCommand(), newCreateCommand(), newPatchCommand(), newDeleteCommand(), newEntitlementsCommand(), newMembersCommand())
	return cmd
}

func newListCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List roles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.RolesAPI.ListRoles(context.TODO()).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			if opts.Filters != "" {
				req = req.Filters(opts.Filters)
			}
			if opts.Sorters != "" {
				req = req.Sorters(opts.Sorters)
			}
			roles, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			rows := make([][]string, 0, len(roles))
			for _, item := range roles {
				rows = append(rows, []string{item.GetName(), item.GetId(), item.GetDescription()})
			}
			return sdkcmd.WriteTable(cmd, []string{"Name", "ID", "Description"}, rows, "Name", roles)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <role-id>",
		Short: "Get a role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			role, resp, err := apiClient.V2024.RolesAPI.GetRole(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, role)
		},
	}
}

func newCreateCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create --file role.json",
		Short: "Create a role from a JSON payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.Role](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			role, resp, err := apiClient.V2024.RolesAPI.CreateRole(context.TODO()).
				Role(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, role)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newPatchCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:     "patch <role-id> --file patch.json",
		Aliases: []string{"update"},
		Short:   "Patch a role from a JSON Patch payload",
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
			role, resp, err := apiClient.V2024.RolesAPI.PatchRole(context.TODO(), args[0]).
				JsonPatchOperation(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, role)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON Patch payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newDeleteCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <role-id>",
		Short: "Delete a role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return clierror.Usage("role delete requires --force")
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			resp, err := apiClient.V2024.RolesAPI.DeleteRole(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, map[string]string{"status": "deleted", "roleId": args[0]})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm role deletion")
	return cmd
}

func newEntitlementsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "entitlements <role-id>",
		Short: "List entitlements for a role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			entitlements, resp, err := apiClient.V2024.RolesAPI.GetRoleEntitlements(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, entitlements)
		},
	}
}

func newMembersCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "members <role-id>",
		Aliases: []string{"identities"},
		Short:   "List identities assigned to a role",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			members, resp, err := apiClient.V2024.RolesAPI.GetRoleAssignedIdentities(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, members)
		},
	}
}
