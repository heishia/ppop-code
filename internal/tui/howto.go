package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HowToModel struct {
	width  int
	height int
	scroll int
}

func NewHowToModel() *HowToModel {
	return &HowToModel{
		scroll: 0,
	}
}

func (m *HowToModel) Init() tea.Cmd {
	return nil
}

func (m *HowToModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Up):
			if m.scroll > 0 {
				m.scroll--
			}
		case key.Matches(msg, DefaultKeyMap.Down):
			m.scroll++
		}
	}

	return m, nil
}

func (m *HowToModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *HowToModel) View() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("📖 How to Start")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Requirements section
	reqTitle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Render("Requirements")
	b.WriteString(reqTitle)
	b.WriteString("\n\n")

	requirements := []string{
		"✓ Claude Code CLI",
		"✓ Cursor CLI",
	}

	for _, line := range requirements {
		b.WriteString(normalStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	notice := mutedStyle.Render("Install both before using ppopcode.")
	b.WriteString(notice)

	// Help
	b.WriteString("\n\n")
	help := helpStyle.Render("esc: back to menu")
	b.WriteString(help)

	content := menuStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
