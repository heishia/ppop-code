package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkflowItem struct {
	name string
	path string
}

func (w WorkflowItem) Title() string       { return w.name }
func (w WorkflowItem) Description() string { return w.path }
func (w WorkflowItem) FilterValue() string { return w.name }

type WorkflowModel struct {
	list     list.Model
	width    int
	height   int
	selected string
	err      error
}

func NewWorkflowModel() *WorkflowModel {
	items := []list.Item{}

	workflowDir := ".vscode/workflows"
	if entries, err := os.ReadDir(workflowDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				name := strings.TrimSuffix(entry.Name(), ".json")
				items = append(items, WorkflowItem{
					name: name,
					path: filepath.Join(workflowDir, entry.Name()),
				})
			}
		}
	}

	if len(items) == 0 {
		items = append(items, WorkflowItem{
			name: "No workflows found",
			path: "Create workflows in .vscode/workflows/",
		})
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

func (m *WorkflowModel) Init() tea.Cmd {
	return nil
}

func (m *WorkflowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Enter):
			if item, ok := m.list.SelectedItem().(WorkflowItem); ok {
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
	Name string
	Path string
	Data map[string]interface{}
}

func (m *WorkflowModel) loadWorkflow(path string) tea.Cmd {
	return func() tea.Msg {
		return WorkflowLoadedMsg{
			Path: path,
		}
	}
}

func (m *WorkflowModel) View() string {
	var b strings.Builder

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
