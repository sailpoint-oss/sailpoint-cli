// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
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
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			var id string

			if len(args) > 0 {
				id = args[0]
			} else {

				transforms, err := GetTransforms()
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
					id = m.Retrieve()[1]
				} else {
					return fmt.Errorf("no transform selected")
				}

			}

			apiClient := config.InitAPIClient()
			_, err := apiClient.V3.TransformsApi.DeleteTransform(context.TODO(), id).Execute()
			if err != nil {
				return err
			}

			err = ListTransforms()
			if err != nil {
				return err
			}

			color.Green("Transform successfully deleted")

			return nil
		},
	}

	return cmd
}
