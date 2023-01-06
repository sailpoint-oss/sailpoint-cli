// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	tuitable "github.com/sailpoint-oss/sailpoint-cli/internal/tui/table"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newDeleteCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [TRANSFORM-ID]",
		Short:   "Delete transform",
		Long:    "Delete a transform",
		Example: "sail transform d 03d5187b-ab96-402c-b5a1-40b74285d77a",
		Aliases: []string{"d"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			endpoint := cmd.Flags().Lookup("transforms-endpoint").Value.String()

			id := ""

			if len(args) > 0 {
				id = args[0]
			}

			if id == "" {

				transforms, err := getTransforms(client, endpoint, cmd)
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
					rows = append(rows, transforms[i].TransformToRows())
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
					return errors.New("no transform selected")
				}

			}

			resp, err := client.Delete(cmd.Context(), util.ResourceUrl(endpoint, id), nil)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("delete transform failed. status: %s\nbody: %s", resp.Status, body)
			}

			err = listTransforms(client, endpoint, cmd)
			if err != nil {
				return err
			}

			color.Green("Transform successfully deleted")

			return nil
		},
	}

	return cmd
}
