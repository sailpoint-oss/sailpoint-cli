package entitlement

import (
	"context"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdkcmd"
	"github.com/spf13/cobra"
)

func NewEntitlementCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entitlement",
		Short:   "Inspect entitlements",
		Long:    "\nInspect Identity Security Cloud entitlements and entitlement hierarchy relationships.\n\n",
		Example: "  sail entitlement list\n  sail entitlement get <entitlement-id>\n  sail entitlement children <entitlement-id>",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(
		newListCommand(),
		newGetCommand(),
		newImportCommand(),
		newParentsCommand(),
		newChildrenCommand(),
	)
	return cmd
}

func newListCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List entitlements",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.EntitlementsAPI.ListEntitlements(context.TODO()).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			if opts.Filters != "" {
				req = req.Filters(opts.Filters)
			}
			if opts.Sorters != "" {
				req = req.Sorters(opts.Sorters)
			}
			entitlements, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			rows := make([][]string, 0, len(entitlements))
			for _, item := range entitlements {
				rows = append(rows, []string{item.GetName(), item.GetId(), item.GetAttribute(), item.GetDescription()})
			}
			return sdkcmd.WriteTable(cmd, []string{"Name", "ID", "Attribute", "Description"}, rows, "Name", entitlements)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <entitlement-id>",
		Short: "Get an entitlement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			entitlement, resp, err := apiClient.V2024.EntitlementsAPI.GetEntitlement(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, entitlement)
		},
	}
}

func newImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import <source-id>",
		Short: "Import entitlements for a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			task, resp, err := apiClient.V2024.EntitlementsAPI.ImportEntitlementsBySource(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, task)
		},
	}
}

func newParentsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "parents <entitlement-id>",
		Short: "List parent entitlements",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			entitlements, resp, err := apiClient.V2024.EntitlementsAPI.ListEntitlementParents(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, entitlements)
		},
	}
}

func newChildrenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "children <entitlement-id>",
		Short: "List child entitlements",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			entitlements, resp, err := apiClient.V2024.EntitlementsAPI.ListEntitlementChildren(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, entitlements)
		},
	}
}
