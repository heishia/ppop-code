package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ppopcode/ppopcode/internal/workflow"
)

// WorkflowItemType distinguishes special items from regular workflow files
type WorkflowItemType int

const (
	WorkflowTypeRegular WorkflowItemType = iota
	WorkflowTypeMakeNew
)

type WorkflowItem struct {
	name     string
	path     string
	itemType WorkflowItemType
}

func (w WorkflowItem) Title() string       { return w.name }
func (w WorkflowItem) Description() string { return w.path }
func (w WorkflowItem) FilterValue() string { return w.name }

type WorkflowModel struct {
	list             list.Model
	width            int
	height           int
	selected         string
	err              error
	showWFStudio     bool // Show WF Studio page
	wfStudioInstall  bool // true = installed, false = not installed
	shortcutLaunched bool // true = user pressed Ctrl+Shift+W
}

// checkWFStudioInstalled checks if cc-wf-studio VSCode/Cursor extension is installed
func checkWFStudioInstalled() bool {
	const extensionID = "breaking-brake.cc-wf-studio"

	// Check Cursor extensions directory
	var extensionDirs []string
	switch runtime.GOOS {
	case "windows":
		home := os.Getenv("USERPROFILE")
		extensionDirs = []string{
			filepath.Join(home, ".cursor", "extensions"),
			filepath.Join(home, ".vscode", "extensions"),
		}
	case "darwin":
		home := os.Getenv("HOME")
		extensionDirs = []string{
			filepath.Join(home, ".cursor", "extensions"),
			filepath.Join(home, ".vscode", "extensions"),
		}
	default:
		home := os.Getenv("HOME")
		extensionDirs = []string{
			filepath.Join(home, ".cursor", "extensions"),
			filepath.Join(home, ".vscode", "extensions"),
		}
	}

	// Check if extension folder exists (e.g., breaking-brake.cc-wf-studio-3.16.1)
	for _, extDir := range extensionDirs {
		if entries, err := os.ReadDir(extDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), extensionID) {
					return true
				}
			}
		}
	}

	return false
}

func NewWorkflowModel() *WorkflowModel {
	items := []list.Item{}

	// Add "Make Workflow" option at the top
	items = append(items, WorkflowItem{
		name:     "✨ Make Workflow",
		path:     "Create new workflow with CC WF Studio",
		itemType: WorkflowTypeMakeNew,
	})

	workflowDir := ".vscode/workflows"
	if entries, err := os.ReadDir(workflowDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				name := strings.TrimSuffix(entry.Name(), ".json")
				items = append(items, WorkflowItem{
					name:     name,
					path:     filepath.Join(workflowDir, entry.Name()),
					itemType: WorkflowTypeRegular,
				})
			}
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.Styles.SelectedDesc = mutedStyle

	l := list.New(items, delegate, 60, 20)
	l.Title = "📋 Workflows"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return &WorkflowModel{
		list: l,
	}
}

func (m *WorkflowModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width-4, height-6)
}

// HasSubView returns true if workflow is showing a sub-view (like WF Studio page)
func (m *WorkflowModel) HasSubView() bool {
	return m.showWFStudio
}

// CloseSubView closes the current sub-view and returns to workflow list
func (m *WorkflowModel) CloseSubView() {
	m.showWFStudio = false
}

// Reset resets the workflow model state (called when entering from menu)
func (m *WorkflowModel) Reset() {
	m.showWFStudio = false
	m.shortcutLaunched = false
	m.selected = ""
}

func (m *WorkflowModel) Init() tea.Cmd {
	return nil
}

func (m *WorkflowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		// Handle WF Studio page navigation
		if m.showWFStudio {
			switch {
			case key.Matches(msg, DefaultKeyMap.Back):
				m.showWFStudio = false
				m.shortcutLaunched = false
				return m, nil
			case key.Matches(msg, DefaultKeyMap.Enter):
				if m.wfStudioInstall {
					// Launch CC WF Studio and show skill paths
					m.shortcutLaunched = true
					return m, m.launchWFStudio()
				} else {
					// Open install page in browser
					return m, m.openInstallPage()
				}
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, DefaultKeyMap.Enter):
			if item, ok := m.list.SelectedItem().(WorkflowItem); ok {
				if item.itemType == WorkflowTypeMakeNew {
					// Check if WF Studio is installed
					m.wfStudioInstall = checkWFStudioInstalled()
					// Always show WF Studio page (for install guide or keybinding setup)
					m.showWFStudio = true
					if m.wfStudioInstall {
						// Register keybinding
						return m, m.launchWFStudio()
					}
					return m, nil
				}
				m.selected = item.path
				return m, m.loadWorkflow(item.path)
			}
		}
	case WorkflowLoadedMsg:
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type WorkflowLoadedMsg struct {
	Name     string
	Path     string
	Workflow *workflow.Workflow
	Error    error
}

type WFStudioLaunchMsg struct {
	Success bool
	Error   error
}

type WFStudioInstallPageMsg struct {
	Success bool
	Error   error
}

func (m *WorkflowModel) loadWorkflow(path string) tea.Cmd {
	return func() tea.Msg {
		// Get the directory and filename
		dir := filepath.Dir(path)
		name := filepath.Base(path)

		// Create loader and load workflow
		loader := workflow.NewLoader(dir)
		wf, err := loader.Load(name)

		if err != nil {
			return WorkflowLoadedMsg{
				Path:  path,
				Error: err,
			}
		}

		return WorkflowLoadedMsg{
			Name:     wf.Name,
			Path:     path,
			Workflow: wf,
		}
	}
}

func (m *WorkflowModel) launchWFStudio() tea.Cmd {
	return func() tea.Msg {
		// Register Ctrl+Shift+W keybinding for CC WF Studio in Cursor user settings
		var keybindingsPath string
		switch runtime.GOOS {
		case "windows":
			keybindingsPath = filepath.Join(os.Getenv("APPDATA"), "Cursor", "User", "keybindings.json")
		case "darwin":
			keybindingsPath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Cursor", "User", "keybindings.json")
		default:
			keybindingsPath = filepath.Join(os.Getenv("HOME"), ".config", "Cursor", "User", "keybindings.json")
		}

		// Read existing keybindings
		var keybindings []map[string]interface{}
		if data, err := os.ReadFile(keybindingsPath); err == nil {
			json.Unmarshal(data, &keybindings)
		}

		// Check if keybinding already exists
		keybindingExists := false
		for _, kb := range keybindings {
			if cmd, ok := kb["command"].(string); ok && cmd == "cc-wf-studio.openEditor" {
				keybindingExists = true
				break
			}
		}

		// Add keybinding if not present
		if !keybindingExists {
			newKeybinding := map[string]interface{}{
				"key":     "ctrl+shift+w",
				"command": "cc-wf-studio.openEditor",
			}
			keybindings = append(keybindings, newKeybinding)

			// Ensure directory exists
			os.MkdirAll(filepath.Dir(keybindingsPath), 0755)

			// Write updated keybindings
			keybindingsData, _ := json.MarshalIndent(keybindings, "", "  ")
			os.WriteFile(keybindingsPath, keybindingsData, 0644)
		}

		return WFStudioLaunchMsg{Success: true, Error: nil}
	}
}

func (m *WorkflowModel) openInstallPage() tea.Cmd {
	return func() tea.Msg {
		url := "https://marketplace.cursorapi.com/items/?itemName=breaking-brake.cc-wf-studio"
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		case "darwin":
			cmd = exec.Command("open", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		err := cmd.Start()
		return WFStudioInstallPageMsg{Success: err == nil, Error: err}
	}
}

func (m *WorkflowModel) View() string {
	var b strings.Builder

	// Show WF Studio page
	if m.showWFStudio {
		return m.viewWFStudio()
	}

	b.WriteString(m.list.View())
	b.WriteString("\n")

	if m.selected != "" {
		selected := fmt.Sprintf("Selected: %s", m.selected)
		b.WriteString(mutedStyle.Render(selected))
		b.WriteString("\n")
	}

	help := helpStyle.Render("↑/↓: navigate • enter: select • /: filter • esc: back")
	b.WriteString(help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m *WorkflowModel) getSkillPathsContent() string {
	var userSkillPath string
	switch runtime.GOOS {
	case "windows":
		userSkillPath = "%USERPROFILE%\\.claude\\skills\\"
	default:
		userSkillPath = "~/.claude/skills/"
	}

	return fmt.Sprintf(`For CC WF Studio's "Browse Skills" panel:

User Skills (all projects):
  %s

Project Skills (current project only):
  .claude/skills/

Add SKILL.md in a subfolder to register.`, userSkillPath)
}

func (m *WorkflowModel) viewWFStudio() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(60)

	if m.wfStudioInstall {
		if m.shortcutLaunched {
			// Second screen: Skill paths only
			b.WriteString(titleStyle.Render("📁 Skill Paths"))
			b.WriteString("\n\n")

			skillPaths := m.getSkillPathsContent()
			b.WriteString(boxStyle.Render(skillPaths))
			b.WriteString("\n\n")

			help := helpStyle.Render("esc: back")
			b.WriteString(help)
		} else {
			// First screen: WF Studio info + shortcut guide
			b.WriteString(titleStyle.Render("🎨 CC WF Studio"))
			b.WriteString("\n\n")

			content := `CC WF Studio extension is installed!

Keybinding registered: Ctrl+Shift+W

Press the shortcut in Cursor to open the Workflow Editor.
(You may need to reload Cursor window once)`

			b.WriteString(boxStyle.Render(content))
			b.WriteString("\n\n")

			// Keybinding highlight
			keybindStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true).
				Background(lipgloss.Color("#333333")).
				Padding(0, 1)
			b.WriteString("Shortcut: ")
			b.WriteString(keybindStyle.Render("Ctrl+Shift+W"))
			b.WriteString("\n\n")

			statusStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)
			b.WriteString(statusStyle.Render("● Installed (Cursor Extension)"))
			b.WriteString("\n\n")

			help := helpStyle.Render("enter: view skill paths • esc: back")
			b.WriteString(help)
		}
	} else {
		// Not installed - show install page
		b.WriteString(titleStyle.Render("🎨 CC WF Studio"))
		b.WriteString("\n\n")

		content := `CC WF Studio extension is not installed.

Visual workflow editor for Claude Code Slash Commands,
Sub Agents, Agent Skills, and MCP Tools.

Install from Cursor Marketplace:
  breaking-brake.cc-wf-studio

Press Enter to open the installation page.`

		b.WriteString(boxStyle.Render(content))
		b.WriteString("\n\n")

		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6600")).
			Bold(true)
		b.WriteString(statusStyle.Render("○ Not Installed"))
		b.WriteString("\n\n")

		help := helpStyle.Render("enter: open Cursor Marketplace • esc: back")
		b.WriteString(help)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}
