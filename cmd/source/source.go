package source

import (
	"context"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdkcmd"
	"github.com/spf13/cobra"
)

func NewSourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "source",
		Short:   "Inspect Identity Security Cloud sources",
		Long:    "\nInspect sources, schemas, connector configuration, provisioning policies, schedules, and health.\n\n",
		Example: "  sail source list\n  sail source get <source-id>\n  sail source schemas <source-id>",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(
		newListCommand(),
		newGetCommand(),
		newCreateCommand(),
		newPatchCommand(),
		newDeleteCommand(),
		newSchemasCommand(),
		newConfigCommand(),
		newHealthCommand(),
		newSchedulesCommand(),
		newPolicyCommand(),
	)
	return cmd
}

func newListCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.SourcesAPI.ListSources(context.TODO()).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			if opts.Filters != "" {
				req = req.Filters(opts.Filters)
			}
			if opts.Sorters != "" {
				req = req.Sorters(opts.Sorters)
			}
			sources, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			rows := make([][]string, 0, len(sources))
			for _, item := range sources {
				rows = append(rows, []string{item.GetName(), item.GetId(), item.GetType(), item.GetStatus()})
			}
			return sdkcmd.WriteTable(cmd, []string{"Name", "ID", "Type", "Status"}, rows, "Name", sources)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <source-id>",
		Short: "Get a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			source, resp, err := apiClient.V2024.SourcesAPI.GetSource(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, source)
		},
	}
}

func newCreateCommand() *cobra.Command {
	var filePath string
	var provisionAsCSV bool
	cmd := &cobra.Command{
		Use:   "create --file source.json",
		Short: "Create a source from a JSON payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.Source](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.SourcesAPI.CreateSource(context.TODO()).Source(payload)
			if provisionAsCSV {
				req = req.ProvisionAsCsv(provisionAsCSV)
			}
			source, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, source)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.Flags().BoolVar(&provisionAsCSV, "provision-as-csv", false, "Create a delimited file source with the provisionAsCsv query parameter")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newPatchCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:     "patch <source-id> --file patch.json",
		Aliases: []string{"update"},
		Short:   "Patch a source from a JSON Patch payload",
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
			source, resp, err := apiClient.V2024.SourcesAPI.UpdateSource(context.TODO(), args[0]).
				JsonPatchOperation(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, source)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON Patch payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newDeleteCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <source-id>",
		Short: "Delete a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return clierror.Usage("source delete requires --force")
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.SourcesAPI.DeleteSource(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm source deletion")
	return cmd
}

func newSchemasCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "schemas <source-id>",
		Short: "List schemas for a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			schemas, resp, err := apiClient.V2024.SourcesAPI.GetSourceSchemas(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, schemas)
		},
	}
}

func newConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config <source-id>",
		Short: "Get connector configuration for a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			cfg, resp, err := apiClient.V2024.SourcesAPI.GetSourceConfig(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, cfg)
		},
	}
}

func newHealthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "health <source-id>",
		Short: "Get source health",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			health, resp, err := apiClient.V2024.SourcesAPI.GetSourceHealth(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, health)
		},
	}
}

func newSchedulesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "schedules <source-id>",
		Short: "List schedules for a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			schedules, resp, err := apiClient.V2024.SourcesAPI.GetSourceSchedules(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, schedules)
		},
	}
}

func newPolicyCommand() *cobra.Command {
	var usageType string
	cmd := &cobra.Command{
		Use:   "policy <source-id>",
		Short: "Get a provisioning policy for a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			policy, resp, err := apiClient.V2024.SourcesAPI.GetProvisioningPolicy(context.TODO(), args[0], v2024.UsageType(usageType)).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, policy)
		},
	}
	cmd.Flags().StringVar(&usageType, "usage-type", "CREATE", "Provisioning policy usage type")
	return cmd
}
