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
	messages      []Message
	input         textarea.Model
	viewport      viewport.Model
	width         int
	height        int
	ready         bool
	processing    bool
	streamingText string // Current streaming output
	thinkingText  string // Current thinking/status text
	currentAgent  string // Current agent processing
	orchestrator  *orchestrator.Orchestrator
	session       *session.Manager
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

			// Handle /clean command
			if content == "/clean" || content == "/clear" {
				m.messages = []Message{}
				m.streamingText = ""
				m.thinkingText = ""
				m.currentAgent = ""
				m.input.Reset()
				m.viewport.SetContent(m.renderMessages())
				return m, nil
			}

			m.messages = append(m.messages, Message{
				Role:    RoleUser,
				Content: content,
			})

			m.input.Reset()
			m.streamingText = ""
			m.thinkingText = ""
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			m.processing = true
			return m, m.sendMessageStream(content)
		}
	case AssistantResponseMsg:
		m.processing = false
		m.streamingText = ""
		m.thinkingText = ""
		m.messages = append(m.messages, Message{
			Role:    RoleAssistant,
			Content: msg.Content,
			Model:   msg.Model,
		})
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case StreamUpdateMsg:
		m.currentAgent = msg.Agent

		switch msg.Type {
		case "status":
			m.thinkingText = msg.Content
		case "thinking":
			m.thinkingText = msg.Content
		case "output":
			m.streamingText += msg.Content
		case "error":
			m.thinkingText = "Error: " + msg.Content
		}

		// Update viewport with current streaming content
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		if msg.Done {
			// Final message received, add to messages
			if m.streamingText != "" {
				m.messages = append(m.messages, Message{
					Role:    RoleAssistant,
					Content: m.streamingText,
					Model:   msg.Agent,
				})
			}
			m.processing = false
			m.streamingText = ""
			m.thinkingText = ""
			m.currentAgent = ""
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			return m, nil
		}

		// Continue listening for more updates
		return m, ListenForStreamUpdates()
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

// StreamUpdateMsg represents a real-time streaming update
type StreamUpdateMsg struct {
	Stage   string // "routing", "processing", "streaming", "completed", "error"
	Content string
	Agent   string
	Type    string // "status", "thinking", "output", "error"
	Done    bool
}

// sendMessageStream sends a message and returns streaming updates
func (m *ChatModel) sendMessageStream(content string) tea.Cmd {
	return func() tea.Msg {
		if m.orchestrator == nil {
			return AssistantResponseMsg{
				Content: fmt.Sprintf("[No orchestrator] Received: %s\n\nPlease configure API keys to enable AI responses.", content),
				Model:   "system",
			}
		}

		// Start streaming process
		ctx := context.Background()
		progressChan := make(chan orchestrator.ProgressUpdate, 100)

		go func() {
			task, _ := m.orchestrator.ProcessStream(ctx, content, progressChan)
			if task != nil && m.session != nil {
				m.session.AddMessage("user", content, "")
				m.session.AddMessage("assistant", task.Result, task.AssignedTo)
			}
		}()

		// Return first update
		update, ok := <-progressChan
		if !ok {
			return StreamUpdateMsg{Done: true}
		}

		// Store channel in closure for subsequent reads
		go m.listenForUpdates(progressChan)

		return StreamUpdateMsg{
			Stage:   update.Stage,
			Content: update.Message,
			Agent:   update.Agent,
			Type:    update.Type,
			Done:    update.Done,
		}
	}
}

// streamUpdateChan is used to send updates to the TUI
var streamUpdateChan = make(chan StreamUpdateMsg, 100)

// listenForUpdates listens for progress updates and sends them to the update channel
func (m *ChatModel) listenForUpdates(progressChan <-chan orchestrator.ProgressUpdate) {
	for update := range progressChan {
		streamUpdateChan <- StreamUpdateMsg{
			Stage:   update.Stage,
			Content: update.Message,
			Agent:   update.Agent,
			Type:    update.Type,
			Done:    update.Done,
		}
	}
}

// ListenForStreamUpdates returns a command that listens for stream updates
func ListenForStreamUpdates() tea.Cmd {
	return func() tea.Msg {
		return <-streamUpdateChan
	}
}

func (m *ChatModel) renderMessages() string {
	if len(m.messages) == 0 && m.streamingText == "" && m.thinkingText == "" {
		welcome := `
Welcome to ppopcode!

I'm your AI orchestrator. I'll analyze your request and route it to the best agent:

  * Gemini 3 Pro  -> UX/UI tasks
  * GPT 5.2       -> Design & debugging
  * Sonnet 4.5    -> General coding

Code modifications will be executed by Cursor.

Commands:
  /clean - Clear chat history

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

	// Show current streaming content
	if m.processing && (m.streamingText != "" || m.thinkingText != "") {
		b.WriteString("\n")

		// Show thinking/status
		if m.thinkingText != "" {
			thinkingStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)
			b.WriteString(thinkingStyle.Render(fmt.Sprintf("[thinking] %s", m.thinkingText)))
			b.WriteString("\n")
		}

		// Show streaming output
		if m.streamingText != "" {
			modelTag := ""
			if m.currentAgent != "" {
				modelTag = mutedStyle.Render(fmt.Sprintf("[%s]", m.currentAgent)) + "\n"
			}
			streamingContent := modelTag + m.streamingText + "_" // Cursor indicator
			bubble := chatBubbleAssistant.Render(streamingContent)
			b.WriteString(bubble)
		}
	}

	return b.String()
}

func (m *ChatModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	header := titleStyle.Render("Chat")

	statusText := ""
	if m.processing {
		if m.currentAgent != "" {
			statusText = lipgloss.NewStyle().Foreground(accentColor).Render(fmt.Sprintf(" [%s] Processing...", m.currentAgent))
		} else {
			statusText = lipgloss.NewStyle().Foreground(accentColor).Render(" Processing...")
		}
	}

	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, header, statusText)

	help := helpStyle.Render("Enter: send | Esc: back to menu | /clean: clear chat | Shift+Enter: newline")

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
