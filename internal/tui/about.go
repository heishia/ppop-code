package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AboutModel struct {
	width  int
	height int
}

func NewAboutModel() *AboutModel {
	return &AboutModel{
		width:  80,
		height: 24,
	}
}

func (m *AboutModel) Init() tea.Cmd {
	return nil
}

func (m *AboutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *AboutModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *AboutModel) View() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("ℹ️ About ppopcode")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Version
	versionLabel := mutedStyle.Render("Version:")
	versionValue := normalStyle.Render(" 1.0.0")
	b.WriteString(versionLabel)
	b.WriteString(versionValue)
	b.WriteString("\n\n")

	// Description
	desc := mutedStyle.Render("Multi-Agent Orchestration TUI")
	b.WriteString(desc)
	b.WriteString("\n")
	desc2 := mutedStyle.Render("Coordinate Claude, Cursor, Gemini, and GPT agents")
	b.WriteString(desc2)
	b.WriteString("\n\n")

	// Links
	linksTitle := selectedStyle.Render("Links")
	b.WriteString(linksTitle)
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("  GitHub: "))
	b.WriteString(normalStyle.Render("https://github.com/ppopcode/ppopcode"))
	b.WriteString("\n\n")

	// Credits
	creditsTitle := selectedStyle.Render("Built with")
	b.WriteString(creditsTitle)
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("  • Bubble Tea (TUI framework)"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("  • Claude Code CLI"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("  • Cursor IDE"))
	b.WriteString("\n\n")

	// Help
	b.WriteString("\n")
	help := helpStyle.Render("esc: back")
	b.WriteString(help)

	content := menuStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
