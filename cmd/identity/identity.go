package identity

import (
	"context"
	"fmt"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdkcmd"
	"github.com/spf13/cobra"
)

func NewIdentityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "identity",
		Short:   "Inspect and operate on identities",
		Long:    "\nInspect Identity Security Cloud identities and related access data.\n\n",
		Example: "  sail identity list\n  sail identity get <identity-id>\n  sail identity entitlements <identity-id>",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(),
		newGetCommand(),
		newEntitlementsCommand(),
		newSyncCommand(),
		newResetCommand(),
		newProcessCommand(),
	)
	return cmd
}

func newListCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List identities",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			req := apiClient.V2024.IdentitiesAPI.ListIdentities(context.TODO()).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			if opts.Filters != "" {
				req = req.Filters(opts.Filters)
			}
			if opts.Sorters != "" {
				req = req.Sorters(opts.Sorters)
			}

			identities, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}

			rows := make([][]string, 0, len(identities))
			for _, item := range identities {
				rows = append(rows, []string{item.GetName(), item.GetId(), item.GetIdentityStatus()})
			}
			return sdkcmd.WriteTable(cmd, []string{"Name", "ID", "Status"}, rows, "Name", identities)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <identity-id>",
		Short: "Get an identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			identity, resp, err := apiClient.V2024.IdentitiesAPI.GetIdentity(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, identity)
		},
	}
}

func newEntitlementsCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	cmd := &cobra.Command{
		Use:   "entitlements <identity-id>",
		Short: "List entitlements for an identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.IdentitiesAPI.ListEntitlementsByIdentity(context.TODO(), args[0]).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			entitlements, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, entitlements)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	cmd.Flags().Lookup("filter").Hidden = true
	cmd.Flags().Lookup("sort").Hidden = true
	return cmd
}

func newSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync <identity-id>",
		Short: "Synchronize attributes for an identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			job, resp, err := apiClient.V2024.IdentitiesAPI.SynchronizeAttributesForIdentity(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, job)
		},
	}
}

func newResetCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "reset <identity-id>",
		Short: "Reset an identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return fmt.Errorf("reset requires --force")
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			resp, err := apiClient.V2024.IdentitiesAPI.ResetIdentity(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, map[string]string{"status": "reset started", "identityId": args[0]})
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm identity reset")
	return cmd
}

func newProcessCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "process --file payload.json",
		Short: "Start identity processing from a JSON payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.ProcessIdentitiesRequest](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			task, resp, err := apiClient.V2024.IdentitiesAPI.StartIdentityProcessing(context.TODO()).
				ProcessIdentitiesRequest(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, task)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}
