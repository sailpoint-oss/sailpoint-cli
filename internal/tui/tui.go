package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type ListItem struct {
	title       string
	description string
}

type Choice struct {
	Title       string
	Description string
}

var choice ListItem

func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string { return i.description }
func (i ListItem) FilterValue() string { return i.title }

type model struct {
	List     list.Model
	Quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.List.SelectedItem().(ListItem)
			if ok {
				choice = i
			}
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.List.View())
}

func (m model) Retrieve() ListItem {
	if choice.title != "" {
		return choice
	} else {
		return ListItem{}
	}
}

func PromptList(choices []Choice, Title string) (Choice, error) {

	items := []list.Item{}
	for i := 0; i < len(choices); i++ {
		entry := choices[i]
		items = append(items, ListItem{title: entry.Title, description: entry.Description})
	}

	m := model{List: list.New(items, list.NewDefaultDelegate(), 0, 0)}

	m.List.Title = Title

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return Choice{}, fmt.Errorf("error running program: %s", err)
	}

	choice := m.Retrieve()

	return Choice{Title: choice.title, Description: choice.description}, nil
}
