package accessrequest

import (
	"context"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/sdkcmd"
	"github.com/spf13/cobra"
)

func NewAccessRequestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "access-request",
		Aliases: []string{"accessrequest"},
		Short:   "Submit and review access requests",
		Long:    "\nSubmit access requests, inspect request status, and act on approver work items.\n\n",
		Example: "  sail access-request list\n  sail access-request create --file request.json\n  sail access-request work-items list",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(
		newListCommand(),
		newGetCommand(),
		newCreateCommand(),
		newCancelCommand(),
		newApproveCommand(),
		newCloseCommand(),
		newWorkItemsCommand(),
	)
	return cmd
}

func newListCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List access request status records",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.AccessRequestsAPI.ListAccessRequestStatus(context.TODO()).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			if opts.Filters != "" {
				req = req.Filters(opts.Filters)
			}
			if opts.Sorters != "" {
				req = req.Sorters(opts.Sorters)
			}
			requests, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			rows := make([][]string, 0, len(requests))
			for _, item := range requests {
				rows = append(rows, []string{item.GetName(), item.GetId(), item.GetType(), string(item.GetState())})
			}
			return sdkcmd.WriteTable(cmd, []string{"Name", "ID", "Type", "State"}, rows, "Name", requests)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <request-id>",
		Short: "Get access request status by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			requests, resp, err := apiClient.V2024.AccessRequestsAPI.ListAccessRequestStatus(context.TODO()).
				Filters(`id eq "` + args[0] + `"`).
				Limit(1).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, requests)
		},
	}
}

func newCreateCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create --file request.json",
		Short: "Create an access request from a JSON payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.AccessRequest](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.AccessRequestsAPI.CreateAccessRequest(context.TODO()).
				AccessRequest(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newCancelCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "cancel --file cancel.json",
		Short: "Cancel an access request from a JSON payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.CancelAccessRequest](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.AccessRequestsAPI.CancelAccessRequest(context.TODO()).
				CancelAccessRequest(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newApproveCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "approve --file approve.json",
		Short: "Approve access requests in bulk from a JSON payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.BulkApproveAccessRequest](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.AccessRequestsAPI.ApproveBulkAccessRequest(context.TODO()).
				BulkApproveAccessRequest(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newCloseCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "close --file close.json",
		Short: "Close an access request from a JSON payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.CloseAccessRequest](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			result, resp, err := apiClient.V2024.AccessRequestsAPI.CloseAccessRequest(context.TODO()).
				CloseAccessRequest(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, result)
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newWorkItemsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "work-items",
		Short: "Inspect and act on access-request work items",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(
		newWorkItemsListCommand(),
		newWorkItemsGetCommand(),
		newWorkItemsApproveCommand(),
		newWorkItemsRejectCommand(),
		newWorkItemsForwardCommand(),
		newWorkItemsCompleteCommand(),
	)
	return cmd
}

func newWorkItemsListCommand() *cobra.Command {
	var opts sdkcmd.ListOptions
	var ownerID string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List work items",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			req := apiClient.V2024.WorkItemsAPI.ListWorkItems(context.TODO()).
				Limit(opts.Limit).
				Offset(opts.Offset).
				Count(opts.Count)
			if ownerID != "" {
				req = req.OwnerId(ownerID)
			}
			items, resp, err := req.Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			rows := make([][]string, 0, len(items))
			for _, item := range items {
				rows = append(rows, []string{item.GetName(), item.GetId(), string(item.GetType()), string(item.GetState())})
			}
			return sdkcmd.WriteTable(cmd, []string{"Name", "ID", "Type", "State"}, rows, "Name", items)
		},
	}
	sdkcmd.AddListFlags(cmd, &opts)
	cmd.Flags().Lookup("filter").Hidden = true
	cmd.Flags().Lookup("sort").Hidden = true
	cmd.Flags().StringVar(&ownerID, "owner-id", "", "Filter work items by owner ID")
	return cmd
}

func newWorkItemsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <work-item-id>",
		Short: "Get a work item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			item, resp, err := apiClient.V2024.WorkItemsAPI.GetWorkItem(context.TODO(), args[0]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, item)
		},
	}
}

func newWorkItemsApproveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <work-item-id> <approval-item-id>",
		Short: "Approve an approval item",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			item, resp, err := apiClient.V2024.WorkItemsAPI.ApproveApprovalItem(context.TODO(), args[0], args[1]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, item)
		},
	}
}

func newWorkItemsRejectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reject <work-item-id> <approval-item-id>",
		Short: "Reject an approval item",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			item, resp, err := apiClient.V2024.WorkItemsAPI.RejectApprovalItem(context.TODO(), args[0], args[1]).Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, item)
		},
	}
}

func newWorkItemsForwardCommand() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "forward <work-item-id> --file forward.json",
		Short: "Forward a work item from a JSON payload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := sdkcmd.ReadJSONFile[v2024.WorkItemForward](filePath)
			if err != nil {
				return err
			}
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			resp, err := apiClient.V2024.WorkItemsAPI.ForwardWorkItem(context.TODO(), args[0]).
				WorkItemForward(payload).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, map[string]string{"status": "forwarded", "workItemId": args[0]})
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "JSON payload file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func newWorkItemsCompleteCommand() *cobra.Command {
	var body string
	cmd := &cobra.Command{
		Use:   "complete <work-item-id> --body value",
		Short: "Complete a work item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}
			item, resp, err := apiClient.V2024.WorkItemsAPI.CompleteWorkItem(context.TODO(), args[0]).
				Body(body).
				Execute()
			if err := sdkcmd.SDKError(resp, err); err != nil {
				return err
			}
			return sdkcmd.WriteStructured(cmd, item)
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "Completion body")
	return cmd
}
