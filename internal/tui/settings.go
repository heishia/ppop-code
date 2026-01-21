package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ppopcode/ppopcode/internal/config"
)

// Available models for each agent type
var claudeModels = []string{
	"claude-sonnet-4-5-20250514",
	"claude-sonnet-4-20250514",
	"claude-opus-4-20250514",
}

var openaiModels = []string{
	"gpt-4o",
	"gpt-4o-mini",
	"gpt-4-turbo",
	"o1",
	"o1-mini",
}

var geminiModels = []string{
	"gemini-2.0-flash",
	"gemini-1.5-pro",
	"gemini-1.5-flash",
}

type AgentSetting struct {
	Name         string
	Type         string
	CurrentModel string
	Models       []string
	ModelIndex   int
}

type SettingsModel struct {
	agents       []AgentSetting
	cursor       int
	width        int
	height       int
	config       *config.Config
	configPath   string
	message      string
	showMessage  bool
	editMode     bool // true when selecting model for current agent
	modelCursor  int  // cursor for model selection
}

// SettingsSavedMsg is sent when settings are saved
type SettingsSavedMsg struct {
	Success bool
	Message string
}

func NewSettingsModel() *SettingsModel {
	return &SettingsModel{
		agents:     []AgentSetting{},
		cursor:     0,
		width:      80,
		height:     24,
		configPath: "config/ppopcode.yaml",
	}
}

func NewSettingsModelWithConfig(cfg *config.Config) *SettingsModel {
	m := NewSettingsModel()
	m.config = cfg
	m.loadAgentsFromConfig()
	return m
}

func (m *SettingsModel) loadAgentsFromConfig() {
	if m.config == nil {
		return
	}

	m.agents = []AgentSetting{}

	// Define order of agents
	agentOrder := []string{"orchestrator", "sonnet", "gemini", "gpt"}

	for _, name := range agentOrder {
		ac, exists := m.config.Agents[name]
		if !exists {
			continue
		}

		var models []string
		switch ac.Type {
		case "claude":
			models = claudeModels
		case "openai":
			models = openaiModels
		case "gemini":
			models = geminiModels
		default:
			continue
		}

		// Find current model index
		modelIndex := 0
		for i, model := range models {
			if model == ac.Model {
				modelIndex = i
				break
			}
		}

		m.agents = append(m.agents, AgentSetting{
			Name:         name,
			Type:         ac.Type,
			CurrentModel: ac.Model,
			Models:       models,
			ModelIndex:   modelIndex,
		})
	}
}

func (m *SettingsModel) Init() tea.Cmd {
	return m.loadConfig
}

func (m *SettingsModel) loadConfig() tea.Msg {
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return SettingsSavedMsg{Success: false, Message: fmt.Sprintf("Failed to load config: %v", err)}
	}
	m.config = cfg
	m.loadAgentsFromConfig()
	return nil
}

func (m *SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case SettingsSavedMsg:
		m.message = msg.Message
		m.showMessage = true
		return m, nil

	case tea.KeyMsg:
		if m.editMode {
			return m.handleEditModeKeys(msg)
		}
		return m.handleNormalModeKeys(msg)
	}

	return m, nil
}

func (m *SettingsModel) handleNormalModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, DefaultKeyMap.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, DefaultKeyMap.Down):
		if m.cursor < len(m.agents)-1 {
			m.cursor++
		}
	case key.Matches(msg, DefaultKeyMap.Enter):
		if m.cursor < len(m.agents) {
			m.editMode = true
			m.modelCursor = m.agents[m.cursor].ModelIndex
		}
	case msg.String() == "s":
		return m, m.saveConfig
	}
	return m, nil
}

func (m *SettingsModel) handleEditModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	agent := &m.agents[m.cursor]

	switch {
	case key.Matches(msg, DefaultKeyMap.Up):
		if m.modelCursor > 0 {
			m.modelCursor--
		}
	case key.Matches(msg, DefaultKeyMap.Down):
		if m.modelCursor < len(agent.Models)-1 {
			m.modelCursor++
		}
	case key.Matches(msg, DefaultKeyMap.Enter):
		// Select model
		agent.ModelIndex = m.modelCursor
		agent.CurrentModel = agent.Models[m.modelCursor]
		m.editMode = false
	case key.Matches(msg, DefaultKeyMap.Back):
		m.editMode = false
	}
	return m, nil
}

func (m *SettingsModel) saveConfig() tea.Msg {
	if m.config == nil {
		return SettingsSavedMsg{Success: false, Message: "No config loaded"}
	}

	// Update config with new models
	for _, agent := range m.agents {
		if ac, exists := m.config.Agents[agent.Name]; exists {
			ac.Model = agent.CurrentModel
			m.config.Agents[agent.Name] = ac
		}
	}

	// Save to file
	if err := m.config.Save(m.configPath); err != nil {
		return SettingsSavedMsg{Success: false, Message: fmt.Sprintf("Failed to save: %v", err)}
	}

	return SettingsSavedMsg{Success: true, Message: "Settings saved!"}
}

func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *SettingsModel) SetConfig(cfg *config.Config) {
	m.config = cfg
	m.loadAgentsFromConfig()
}

func (m *SettingsModel) View() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("⚙️  Settings")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Subtitle
	subtitle := mutedStyle.Render("Configure AI models for each agent")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	if len(m.agents) == 0 {
		b.WriteString(mutedStyle.Render("Loading configuration..."))
		b.WriteString("\n")
	} else if m.editMode {
		// Model selection mode
		agent := m.agents[m.cursor]
		b.WriteString(selectedStyle.Render(fmt.Sprintf("Select model for %s:", agent.Name)))
		b.WriteString("\n\n")

		for i, model := range agent.Models {
			cursor := "  "
			style := normalStyle

			if i == m.modelCursor {
				cursor = "▸ "
				style = selectedStyle
			}

			// Mark current model
			current := ""
			if model == agent.CurrentModel {
				current = " (current)"
			}

			b.WriteString(style.Render(fmt.Sprintf("%s%s%s", cursor, model, current)))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		help := helpStyle.Render("↑/↓: navigate • enter: select • esc: cancel")
		b.WriteString(help)
	} else {
		// Agent list mode
		for i, agent := range m.agents {
			cursor := "  "
			style := normalStyle

			if i == m.cursor {
				cursor = "▸ "
				style = selectedStyle
			}

			// Agent type icon
			icon := "🤖"
			switch agent.Type {
			case "claude":
				icon = "🟣"
			case "openai":
				icon = "🟢"
			case "gemini":
				icon = "🔵"
			}

			// Title line
			titleLine := fmt.Sprintf("%s%s %s", cursor, icon, agent.Name)
			b.WriteString(style.Render(titleLine))
			b.WriteString("\n")

			// Model line
			modelLine := fmt.Sprintf("    Model: %s", agent.CurrentModel)
			if i == m.cursor {
				b.WriteString(accentStyle.Render(modelLine))
			} else {
				b.WriteString(mutedStyle.Render(modelLine))
			}
			b.WriteString("\n\n")
		}

		// Message
		if m.showMessage && m.message != "" {
			msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
			b.WriteString(msgStyle.Render(m.message))
			b.WriteString("\n\n")
		}

		// Help
		help := helpStyle.Render("↑/↓: navigate • enter: change model • s: save • esc: back")
		b.WriteString(help)
	}

	content := menuStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
