package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuItem struct {
	Title       string
	Description string
	Icon        string
}

type MenuModel struct {
	items    []MenuItem
	cursor   int
	Selected int
	width    int
	height   int
}

func NewMenuModel() *MenuModel {
	return &MenuModel{
		items: []MenuItem{
			{
				Title:       "Link Accounts",
				Description: "Setup Claude and Cursor authentication",
				Icon:        "🔗",
			},
			{
				Title:       "Chat",
				Description: "Start a conversation with AI agents",
				Icon:        "💬",
			},
			{
				Title:       "Workflow",
				Description: "Select and run cc-wf-studio workflows",
				Icon:        "📋",
			},
			{
				Title:       "How to Start",
				Description: "Learn how to setup and use ppopcode",
				Icon:        "📖",
			},
			{
				Title:       "Settings",
				Description: "Configure agents and preferences",
				Icon:        "⚙️",
			},
		},
		cursor:   0,
		Selected: -1,
	}
}

func (m *MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, DefaultKeyMap.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case key.Matches(msg, DefaultKeyMap.Enter):
			m.Selected = m.cursor
			return m, nil
		}
	}

	return m, nil
}

func (m *MenuModel) View() string {
	var b strings.Builder

	logo := `
    ____  ____  ____  ____  _____ ____  ____  ______
   / __ \/ __ \/ __ \/ __ \/ ___// __ \/ __ \/ ____/
  / /_/ / /_/ / / / / /_/ / /   / / / / / / / __/   
 / ____/ ____/ /_/ / ____/ /___/ /_/ / /_/ / /___   
/_/   /_/    \____/_/    \____/\____/_____/_____/   
`

	logoStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	b.WriteString(logoStyle.Render(logo))
	b.WriteString("\n")

	subtitle := mutedStyle.Render("Multi-Agent Orchestration TUI")
	b.WriteString(lipgloss.PlaceHorizontal(60, lipgloss.Center, subtitle))
	b.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := normalStyle

		if i == m.cursor {
			cursor = "▸ "
			style = selectedStyle
		}

		itemStr := fmt.Sprintf("%s%s %s", cursor, item.Icon, item.Title)
		b.WriteString(style.Render(itemStr))
		b.WriteString("\n")

		if i == m.cursor {
			desc := fmt.Sprintf("    %s", item.Description)
			b.WriteString(mutedStyle.Render(desc))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	help := helpStyle.Render("↑/↓: navigate • enter: select • q: quit")
	b.WriteString(help)

	content := menuStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
