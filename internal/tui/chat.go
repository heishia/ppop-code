package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ppopcode/ppopcode/internal/orchestrator"
	"github.com/ppopcode/ppopcode/internal/session"
)

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
)

type Message struct {
	Role    MessageRole
	Content string
	Model   string
}

type ChatModel struct {
	messages     []Message
	input        textarea.Model
	viewport     viewport.Model
	width        int
	height       int
	ready        bool
	processing   bool
	orchestrator *orchestrator.Orchestrator
	session      *session.Manager
}

func NewChatModel() *ChatModel {
	ti := textarea.New()
	ti.Placeholder = "Type your message... (Enter to send, Shift+Enter for newline)"
	ti.CharLimit = 4096
	ti.SetWidth(80)
	ti.SetHeight(3)
	ti.ShowLineNumbers = false
	ti.Focus()

	return &ChatModel{
		messages:   []Message{},
		input:      ti,
		processing: false,
	}
}

func NewChatModelWithOrchestrator(orch *orchestrator.Orchestrator, sess *session.Manager) *ChatModel {
	ti := textarea.New()
	ti.Placeholder = "Type your message... (Enter to send, Shift+Enter for newline)"
	ti.CharLimit = 4096
	ti.SetWidth(80)
	ti.SetHeight(3)
	ti.ShowLineNumbers = false
	ti.Focus()

	return &ChatModel{
		messages:     []Message{},
		input:        ti,
		processing:   false,
		orchestrator: orch,
		session:      sess,
	}
}

func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	headerHeight := 3
	inputHeight := 5
	viewportHeight := height - headerHeight - inputHeight - 2

	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.viewport = viewport.New(width-4, viewportHeight)
	m.viewport.SetContent(m.renderMessages())
	m.input.SetWidth(width - 4)
	m.ready = true
}

func (m *ChatModel) Focus() {
	m.input.Focus()
}

func (m *ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Back):
			return m, nil

		case msg.Type == tea.KeyEnter && !msg.Alt:
			if m.processing {
				return m, nil
			}

			content := strings.TrimSpace(m.input.Value())
			if content == "" {
				return m, nil
			}

			m.messages = append(m.messages, Message{
				Role:    RoleUser,
				Content: content,
			})

			m.input.Reset()
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			m.processing = true
			return m, m.sendMessage(content)
		}
	case AssistantResponseMsg:
		m.processing = false
		m.messages = append(m.messages, Message{
			Role:    RoleAssistant,
			Content: msg.Content,
			Model:   msg.Model,
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil
	}

	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

type AssistantResponseMsg struct {
	Content string
	Model   string
}

func (m *ChatModel) sendMessage(content string) tea.Cmd {
	return func() tea.Msg {
		if m.orchestrator == nil {
			return AssistantResponseMsg{
				Content: fmt.Sprintf("[No orchestrator] Received: %s\n\nPlease configure API keys to enable AI responses.", content),
				Model:   "system",
			}
		}

		ctx := context.Background()
		task, err := m.orchestrator.Process(ctx, content)
		if err != nil {
			return AssistantResponseMsg{
				Content: fmt.Sprintf("Error: %v", err),
				Model:   "error",
			}
		}

		if m.session != nil {
			m.session.AddMessage("user", content, "")
			m.session.AddMessage("assistant", task.Result, task.AssignedTo)
		}

		return AssistantResponseMsg{
			Content: task.Result,
			Model:   task.AssignedTo,
		}
	}
}

func (m *ChatModel) renderMessages() string {
	if len(m.messages) == 0 {
		welcome := `
Welcome to ppopcode!

I'm your AI orchestrator. I'll analyze your request and route it to the best agent:

  • Gemini 3 Pro  → UX/UI tasks
  • GPT 5.2       → Design & debugging
  • Sonnet 4.5    → General coding

Code modifications will be executed by Cursor.

How can I help you today?
`
		return mutedStyle.Render(welcome)
	}

	var b strings.Builder
	for _, msg := range m.messages {
		switch msg.Role {
		case RoleUser:
			bubble := chatBubbleUser.Render(msg.Content)
			b.WriteString(lipgloss.PlaceHorizontal(m.width-4, lipgloss.Right, bubble))
		case RoleAssistant:
			modelTag := ""
			if msg.Model != "" {
				modelTag = mutedStyle.Render(fmt.Sprintf("[%s]", msg.Model)) + "\n"
			}
			bubble := chatBubbleAssistant.Render(modelTag + msg.Content)
			b.WriteString(bubble)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *ChatModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	header := titleStyle.Render("💬 Chat")

	statusText := ""
	if m.processing {
		statusText = lipgloss.NewStyle().Foreground(accentColor).Render(" ⏳ Processing...")
	}

	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, header, statusText)

	help := helpStyle.Render("Enter: send • Esc: back to menu • Shift+Enter: newline")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine,
		"",
		m.viewport.View(),
		"",
		inputStyle.Render(m.input.View()),
		help,
	)
}
