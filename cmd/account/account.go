package account

import (
	"context"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdkcmd"
	"github.com/spf13/cobra"
)

func NewAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account",
		Short:   "Inspect and operate on accounts",
		Long:    "\nInspect accounts and perform common account operations such as enable, disable, unlock, and reload.\n\n",
		Example: "  sail account list\n  sail account get <account-id>\n  sail account disable <account-id>",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(
		newListCommand(),
		newGetCommand(),
		newEntitlementsCommand(),
		newEnableCommand(),
		newDisableCommand(),
		newUnlockCommand(),
		newReloadCommand(),
	)
	return cmd
}

func newListCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.AccountsAPI.ListAccounts(context.TODO()).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			if opts.Filters != "" {
				req = req.Filters(opts.Filters)
			}
			if opts.Sorters != "" {
				req = req.Sorters(opts.Sorters)
			}
			accounts, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			rows := make([][]string, 0, len(accounts))
			for _, item := range accounts {
				rows = append(rows, []string{item.GetName(), item.GetId(), item.GetSourceName(), item.GetNativeIdentity()})
			}
			return sdkcmd.WriteTable(cmd, []string{"Name", "ID", "Source", "Native Identity"}, rows, "Name", accounts)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <account-id>",
		Short: "Get an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			account, resp, err := apiClient.V2024.AccountsAPI.GetAccount(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, account)
		},
	}
}

func newEntitlementsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "entitlements <account-id>",
		Short: "List entitlements for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			entitlements, resp, err := apiClient.V2024.AccountsAPI.GetAccountEntitlements(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, entitlements)
		},
	}
}

func newEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <account-id>",
		Short: "Enable an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.AccountsAPI.EnableAccount(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
}

func newDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <account-id>",
		Short: "Disable an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.AccountsAPI.DisableAccount(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
}

func newUnlockCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unlock <account-id>",
		Short: "Unlock an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.AccountsAPI.UnlockAccount(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
}

func newReloadCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reload <account-id>",
		Short: "Reload an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.AccountsAPI.SubmitReloadAccount(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
}
