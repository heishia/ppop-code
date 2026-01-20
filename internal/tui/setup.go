package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ppopcode/ppopcode/internal/agents"
	"github.com/ppopcode/ppopcode/internal/cursor"
)

type SetupItem struct {
	Title      string
	Status     string
	StatusOK   bool
	Action     string
	ActionFunc func() error
}

type SetupModel struct {
	items   []SetupItem
	cursor  int
	width   int
	height  int
	message string
	loading bool
}

// StatusCheckMsg is sent when status check completes
type StatusCheckMsg struct {
	claudeStatus agents.ClaudeLoginStatus
	cursorStatus cursor.CursorLoginStatus
}

// ActionResultMsg is sent when an action completes
type ActionResultMsg struct {
	success bool
	message string
}

func NewSetupModel() *SetupModel {
	return &SetupModel{
		items:   []SetupItem{},
		cursor:  0,
		loading: true,
	}
}

func (m *SetupModel) Init() tea.Cmd {
	return m.checkStatus
}

func (m *SetupModel) checkStatus() tea.Msg {
	claudeStatus := agents.CheckClaudeLogin()
	cursorStatus := cursor.CheckCursorLogin()

	return StatusCheckMsg{
		claudeStatus: claudeStatus,
		cursorStatus: cursorStatus,
	}
}

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case StatusCheckMsg:
		m.loading = false
		m.items = m.buildItems(msg.claudeStatus, msg.cursorStatus)
		return m, nil

	case ActionResultMsg:
		m.message = msg.message
		// Refresh status after action
		return m, m.checkStatus

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
			if m.cursor < len(m.items) && m.items[m.cursor].ActionFunc != nil {
				return m, m.runAction(m.cursor)
			}
		case msg.String() == "r":
			// Refresh status
			m.loading = true
			return m, m.checkStatus
		}
	}

	return m, nil
}

func (m *SetupModel) buildItems(claudeStatus agents.ClaudeLoginStatus, cursorStatus cursor.CursorLoginStatus) []SetupItem {
	items := []SetupItem{}

	// Claude item
	claudeItem := SetupItem{
		Title: "Claude Code",
	}
	if claudeStatus.CLIFound {
		if claudeStatus.LoggedIn {
			claudeItem.Status = "Ready"
			claudeItem.StatusOK = true
			claudeItem.Action = "Logged in"
		} else {
			claudeItem.Status = "Not logged in"
			claudeItem.StatusOK = false
			claudeItem.Action = "Run 'claude login'"
			claudeItem.ActionFunc = func() error {
				return agents.RunClaudeLogin()
			}
		}
	} else {
		claudeItem.Status = "CLI not found"
		claudeItem.StatusOK = false
		claudeItem.Action = "Install Claude CLI"
	}
	items = append(items, claudeItem)

	// Cursor item
	cursorItem := SetupItem{
		Title: "Cursor",
	}
	if cursorStatus.Available {
		cursorItem.Status = "Ready"
		cursorItem.StatusOK = true
		cursorItem.Action = cursorStatus.Message
	} else {
		cursorItem.Status = "Not found"
		cursorItem.StatusOK = false
		cursorItem.Action = "Open Cursor IDE"
		cursorItem.ActionFunc = func() error {
			return cursor.OpenCursorLogin()
		}
	}
	items = append(items, cursorItem)

	return items
}

func (m *SetupModel) runAction(index int) tea.Cmd {
	return func() tea.Msg {
		if index >= len(m.items) || m.items[index].ActionFunc == nil {
			return ActionResultMsg{
				success: false,
				message: "No action available",
			}
		}

		err := m.items[index].ActionFunc()
		if err != nil {
			return ActionResultMsg{
				success: false,
				message: fmt.Sprintf("Error: %v", err),
			}
		}

		return ActionResultMsg{
			success: true,
			message: "Action started. Please complete the setup in the opened window.",
		}
	}
}

func (m *SetupModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *SetupModel) View() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("🔗 Link Accounts")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Subtitle
	subtitle := mutedStyle.Render("Setup your authentication to use ppopcode")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(mutedStyle.Render("Checking status..."))
		b.WriteString("\n")
	} else {
		// Status items
		for i, item := range m.items {
			cursor := "  "
			style := normalStyle

			if i == m.cursor {
				cursor = "▸ "
				style = selectedStyle
			}

			// Status indicator
			statusIcon := "❌"
			statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
			if item.StatusOK {
				statusIcon = "✅"
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
			}

			// Title line
			titleLine := fmt.Sprintf("%s%s", cursor, item.Title)
			b.WriteString(style.Render(titleLine))
			b.WriteString("\n")

			// Status line
			statusLine := fmt.Sprintf("    %s Status: %s", statusIcon, item.Status)
			b.WriteString(statusStyle.Render(statusLine))
			b.WriteString("\n")

			// Action line
			actionLine := fmt.Sprintf("    → %s", item.Action)
			if item.ActionFunc != nil && i == m.cursor {
				b.WriteString(accentStyle.Render(actionLine))
			} else {
				b.WriteString(mutedStyle.Render(actionLine))
			}
			b.WriteString("\n\n")
		}
	}

	// Message
	if m.message != "" {
		b.WriteString("\n")
		b.WriteString(accentStyle.Render(m.message))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	help := helpStyle.Render("↑/↓: navigate • enter: run action • r: refresh • esc: back")
	b.WriteString(help)

	content := menuStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

var accentStyle = lipgloss.NewStyle().
	Foreground(accentColor).
	Bold(true)
