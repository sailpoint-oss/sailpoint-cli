// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete an IdentityNow transform",
		Long:    "\nDelete an IdentityNow transform\n\n",
		Example: "sail transform delete 03d5187b-ab96-402c-b5a1-40b74285d77a",
		Aliases: []string{"d"},
		RunE: func(cmd *cobra.Command, args []string) error {

			apiClient, err := config.InitAPIClient()
			if err != nil {
				return err
			}

			for _, transformID := range args {

				_, err = apiClient.V3.TransformsApi.DeleteTransform(context.TODO(), transformID).Execute()
				if err != nil {
					return err
				}

				log.Info("Transform successfully deleted", "TransformID", transformID)
			}

			return nil
		},
	}

	return cmd
}
