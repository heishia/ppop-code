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

	// Prerequisites section
	prereqTitle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Render("Prerequisites")
	b.WriteString(prereqTitle)
	b.WriteString("\n\n")

	prerequisites := []string{
		"1. Claude Account",
		"   → Sign up at https://claude.ai",
		"   → Required for Claude Code CLI",
		"",
		"2. Claude Code CLI",
		"   → Install: npm install -g @anthropic-ai/claude-code",
		"   → Login:   claude login",
		"   → This opens browser for authentication",
		"",
		"3. Cursor IDE (Recommended)",
		"   → Download from https://cursor.sh",
		"   → Used for AI-powered code editing",
		"",
	}

	for _, line := range prerequisites {
		if strings.HasPrefix(line, "   →") {
			b.WriteString(mutedStyle.Render(line))
		} else if line == "" {
			b.WriteString("")
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Quick Start section
	quickTitle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Render("Quick Start")
	b.WriteString("\n")
	b.WriteString(quickTitle)
	b.WriteString("\n\n")

	quickStart := []string{
		"Step 1: Install Claude Code CLI",
		"   $ npm install -g @anthropic-ai/claude-code",
		"",
		"Step 2: Login to Claude",
		"   $ claude login",
		"   (Browser will open for authentication)",
		"",
		"Step 3: Run ppopcode",
		"   $ ppopcode",
		"",
		"Step 4: Go to 'Link Accounts' to verify setup",
		"   Both Claude Code and Cursor should show ✅",
		"",
	}

	for _, line := range quickStart {
		if strings.HasPrefix(line, "   $") {
			cmdStyle := lipgloss.NewStyle().Foreground(accentColor)
			b.WriteString(cmdStyle.Render(line))
		} else if strings.HasPrefix(line, "   ") {
			b.WriteString(mutedStyle.Render(line))
		} else if line == "" {
			b.WriteString("")
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Features section
	featTitle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Render("Features")
	b.WriteString("\n")
	b.WriteString(featTitle)
	b.WriteString("\n\n")

	features := []string{
		"🔗 Link Accounts - Setup and verify authentication",
		"💬 Chat         - Talk with AI agents (Claude, GPT, Gemini)",
		"📋 Workflow     - Run automated coding workflows",
		"⚙️  Settings     - Configure agents and preferences",
	}

	for _, line := range features {
		b.WriteString(normalStyle.Render(line))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	help := helpStyle.Render("↑/↓: scroll • esc: back to menu")
	b.WriteString(help)

	content := menuStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
