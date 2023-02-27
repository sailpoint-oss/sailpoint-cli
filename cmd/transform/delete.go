// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/transform"
	tuitable "github.com/sailpoint-oss/sailpoint-cli/internal/tui/table"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [TRANSFORM-ID]",
		Short:   "delete transform",
		Long:    "Delete a transform",
		Example: "sail transform d 03d5187b-ab96-402c-b5a1-40b74285d77a",
		Aliases: []string{"d"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var id []string

			if len(args) < 1 {
				transforms, err := transform.GetTransforms()
				if err != nil {
					return err
				}

				columns := []table.Column{
					{Title: "Name", Width: 25},
					{Title: "ID", Width: 40},
					{Title: "Type", Width: 25},
				}

				var rows []table.Row

				for i := 0; i < len(transforms); i++ {
					transform := transforms[i]
					rows = append(rows, []string{*transform.Id, transform.Name})
				}

				t := table.New(
					table.WithColumns(columns),
					table.WithRows(rows),
					table.WithFocused(true),
					table.WithHeight(7),
				)

				s := table.DefaultStyles()
				s.Header = s.Header.
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color("240")).
					BorderBottom(true).
					Bold(false)
				s.Selected = s.Selected.
					Foreground(lipgloss.Color("229")).
					Background(lipgloss.Color("57")).
					Bold(false)
				t.SetStyles(s)

				m := tuitable.Model{Table: t}
				if _, err := tea.NewProgram(m).Run(); err != nil {
					fmt.Println("Error running program:", err)
					os.Exit(1)
				}

				tempRow := m.Retrieve()

				if len(tempRow) > 0 {
					id = append(id, m.Retrieve()[1])
				} else {
					return fmt.Errorf("no transform selected")
				}
			} else {
				id = args
			}

			for i := 0; i < len(id); i++ {

				transformID := id[i]

				apiClient, err := config.InitAPIClient()
				if err != nil {
					return err
				}

				_, err = apiClient.V3.TransformsApi.DeleteTransform(context.TODO(), transformID).Execute()
				if err != nil {
					return err
				}

				log.Log.Info("Transform successfully deleted", "TransformID", transformID)
			}

			err := transform.ListTransforms()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
