package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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
	streamingText string
	thinkingText  string
	currentAgent  string
	orchestrator  *orchestrator.Orchestrator
	session       *session.Manager
	progressChan  <-chan orchestrator.ProgressUpdate // Active progress channel
	spinner       spinner.Model
	startTime     time.Time
}

func NewChatModel() *ChatModel {
	ti := textarea.New()
	ti.Placeholder = "Type your message... (Enter to send, Shift+Enter for newline)"
	ti.CharLimit = 4096
	ti.SetWidth(80)
	ti.SetHeight(3)
	ti.ShowLineNumbers = false
	ti.Focus()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)

	return &ChatModel{
		messages:   []Message{},
		input:      ti,
		processing: false,
		spinner:    s,
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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)

	return &ChatModel{
		messages:     []Message{},
		input:        ti,
		processing:   false,
		orchestrator: orch,
		session:      sess,
		spinner:      s,
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

// StreamUpdateMsg represents a real-time streaming update
type StreamUpdateMsg struct {
	Content string
	Agent   string
	Type    string // "status", "thinking", "output", "error"
	Done    bool
}

// tickMsg is sent periodically to update spinner and elapsed time
type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case tickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, tea.Batch(cmd, tickCmd())
		}
		return m, nil

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
			m.startTime = time.Now()

			// Start streaming and get the channel
			progressChan := m.startStreaming(content)
			m.progressChan = progressChan

			// Return command to listen for first update and start spinner tick
			return m, tea.Batch(m.waitForUpdate(), tickCmd())
		}

	case StreamUpdateMsg:
		if msg.Agent != "" {
			m.currentAgent = msg.Agent
		}

		switch msg.Type {
		case "status":
			m.thinkingText = msg.Content
		case "thinking":
			m.thinkingText = msg.Content
		case "output":
			m.streamingText += msg.Content
			m.thinkingText = "" // Clear thinking when output starts
		case "error":
			m.thinkingText = "Error: " + msg.Content
		}

		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

		if msg.Done {
			if m.streamingText != "" {
				m.messages = append(m.messages, Message{
					Role:    RoleAssistant,
					Content: m.streamingText,
					Model:   m.currentAgent,
				})
			}
			m.processing = false
			m.streamingText = ""
			m.thinkingText = ""
			m.currentAgent = ""
			m.progressChan = nil
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
			return m, nil
		}

		// Continue listening for more updates
		return m, m.waitForUpdate()
	}

	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

// startStreaming starts the streaming process and returns the progress channel
// Channel is now owned and managed by Orchestrator
func (m *ChatModel) startStreaming(content string) <-chan orchestrator.ProgressUpdate {
	if m.orchestrator == nil {
		ch := make(chan orchestrator.ProgressUpdate, 1)
		ch <- orchestrator.ProgressUpdate{
			Message: "[No orchestrator] Please configure API keys.",
			Type:    "error",
			Done:    true,
		}
		close(ch)
		return ch
	}

	ctx := context.Background()
	return m.orchestrator.ProcessStreamAsync(ctx, content)
}

// waitForUpdate returns a command that waits for the next progress update
func (m *ChatModel) waitForUpdate() tea.Cmd {
	ch := m.progressChan
	if ch == nil {
		return nil
	}

	return func() tea.Msg {
		update, ok := <-ch
		if !ok {
			return StreamUpdateMsg{Done: true}
		}
		return StreamUpdateMsg{
			Content: update.Message,
			Agent:   update.Agent,
			Type:    update.Type,
			Done:    update.Done,
		}
	}
}

type AssistantResponseMsg struct {
	Content string
	Model   string
}

func (m *ChatModel) renderMessages() string {
	if len(m.messages) == 0 && m.streamingText == "" && m.thinkingText == "" {
		welcome := `
Welcome to ppopcode!

I'm your AI coding assistant powered by Claude.
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
	if m.processing {
		b.WriteString("\n")

		// Show thinking/status
		if m.thinkingText != "" {
			thinkingStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("243")).
				Italic(true)
			b.WriteString(thinkingStyle.Render(fmt.Sprintf("[%s] %s", m.currentAgent, m.thinkingText)))
			b.WriteString("\n")
		}

		// Show streaming output
		if m.streamingText != "" {
			modelTag := ""
			if m.currentAgent != "" {
				modelTag = mutedStyle.Render(fmt.Sprintf("[%s]", m.currentAgent)) + "\n"
			}
			streamingContent := modelTag + m.streamingText + "_"
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
		elapsed := time.Since(m.startTime)
		elapsedStr := fmt.Sprintf("(%ds)", int(elapsed.Seconds()))
		statusStyle := lipgloss.NewStyle().Foreground(accentColor)

		if m.thinkingText != "" {
			statusText = statusStyle.Render(fmt.Sprintf(" %s [%s] %s %s", m.spinner.View(), m.currentAgent, m.thinkingText, elapsedStr))
		} else if m.currentAgent != "" {
			statusText = statusStyle.Render(fmt.Sprintf(" %s [%s] Processing... %s", m.spinner.View(), m.currentAgent, elapsedStr))
		} else {
			statusText = statusStyle.Render(fmt.Sprintf(" %s Initializing... %s", m.spinner.View(), elapsedStr))
		}
	}

	headerLine := lipgloss.JoinHorizontal(lipgloss.Left, header, statusText)

	help := helpStyle.Render("Enter: send | Esc: back | /clean: clear | Shift+Enter: newline")

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
