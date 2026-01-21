package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ppopcode/ppopcode/internal/orchestrator"
	"github.com/ppopcode/ppopcode/internal/workflow"
)

// NodeStatus represents the execution status of a node
type NodeStatus int

const (
	NodePending NodeStatus = iota
	NodeRunning
	NodeCompleted
	NodeError
	NodeWaitingInput
)

// NodeDisplayItem represents a node in the display list
type NodeDisplayItem struct {
	ID     string
	Name   string
	Type   string
	Status NodeStatus
	Output string
}

// WorkflowRunModel handles the workflow execution UI
type WorkflowRunModel struct {
	workflow     *workflow.Workflow
	executor     *workflow.Executor
	nodes        []NodeDisplayItem
	currentNode  int
	output       strings.Builder
	progressChan <-chan workflow.ExecutionProgress
	ctx          context.Context
	cancel       context.CancelFunc

	// UI components
	viewport viewport.Model
	spinner  spinner.Model
	width    int
	height   int
	ready    bool

	// Execution state
	running   bool
	completed bool
	errMsg    string
	startTime time.Time

	// askUserQuestion support
	waitingInput bool
	inputField   textarea.Model
	questionText string
	options      []string
}

// ExecutionProgressMsg wraps workflow.ExecutionProgress for the TUI
type ExecutionProgressMsg struct {
	workflow.ExecutionProgress
}

// workflowTickMsg is sent periodically to update spinner
type workflowTickMsg time.Time

// NewWorkflowRunModel creates a new workflow run model
func NewWorkflowRunModel(wf *workflow.Workflow, orch *orchestrator.Orchestrator) *WorkflowRunModel {
	// Create executor
	executor := workflow.NewExecutor(wf, orch)

	// Build node display list from workflow
	nodes := buildNodeDisplayList(wf)

	// Setup spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)

	// Setup input field for askUserQuestion
	ti := textarea.New()
	ti.Placeholder = "Type your answer..."
	ti.CharLimit = 1024
	ti.SetWidth(60)
	ti.SetHeight(2)
	ti.ShowLineNumbers = false

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkflowRunModel{
		workflow:   wf,
		executor:   executor,
		nodes:      nodes,
		spinner:    s,
		inputField: ti,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// buildNodeDisplayList creates a display list from workflow nodes
func buildNodeDisplayList(wf *workflow.Workflow) []NodeDisplayItem {
	var items []NodeDisplayItem

	// Get execution order by traversing from start
	visited := make(map[string]bool)
	var traverse func(nodeID string)
	traverse = func(nodeID string) {
		if visited[nodeID] {
			return
		}
		node := wf.GetNode(nodeID)
		if node == nil {
			return
		}
		visited[nodeID] = true

		name := node.Data.Label
		if name == "" {
			name = node.Type
		}

		items = append(items, NodeDisplayItem{
			ID:     node.ID,
			Name:   name,
			Type:   node.Type,
			Status: NodePending,
		})

		nextNodes := wf.GetNextNodes(nodeID)
		for _, next := range nextNodes {
			traverse(next.ID)
		}
	}

	startNode := wf.GetStartNode()
	if startNode != nil {
		traverse(startNode.ID)
	}

	return items
}

func (m *WorkflowRunModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate viewport size (right panel for output)
	nodeListWidth := 30
	outputWidth := width - nodeListWidth - 6

	if outputWidth < 20 {
		outputWidth = 20
	}

	m.viewport = viewport.New(outputWidth, height-10)
	m.viewport.SetContent(m.output.String())
	m.inputField.SetWidth(outputWidth - 4)
	m.ready = true
}

func (m *WorkflowRunModel) Init() tea.Cmd {
	// Start execution
	m.running = true
	m.startTime = time.Now()
	m.progressChan = m.executor.ExecuteAsync(m.ctx)

	return tea.Batch(
		m.waitForProgress(),
		workflowTickCmd(),
	)
}

func workflowTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return workflowTickMsg(t)
	})
}

func (m *WorkflowRunModel) waitForProgress() tea.Cmd {
	ch := m.progressChan
	if ch == nil {
		return nil
	}

	return func() tea.Msg {
		progress, ok := <-ch
		if !ok {
			return ExecutionProgressMsg{
				ExecutionProgress: workflow.ExecutionProgress{Done: true},
			}
		}
		return ExecutionProgressMsg{ExecutionProgress: progress}
	}
}

func (m *WorkflowRunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case workflowTickMsg:
		if m.running {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, tea.Batch(cmd, workflowTickCmd())
		}
		return m, nil

	case ExecutionProgressMsg:
		return m.handleProgress(msg)

	case tea.KeyMsg:
		// Handle user input for askUserQuestion
		if m.waitingInput {
			switch msg.Type {
			case tea.KeyEnter:
				if !msg.Alt {
					answer := strings.TrimSpace(m.inputField.Value())
					if answer != "" {
						m.executor.ProvideAnswer(answer)
						m.waitingInput = false
						m.inputField.Reset()

						// Update node status
						for i := range m.nodes {
							if m.nodes[i].Status == NodeWaitingInput {
								m.nodes[i].Status = NodeCompleted
								m.nodes[i].Output = "Answer: " + answer
								break
							}
						}

						m.output.WriteString(fmt.Sprintf("\n[Answer] %s\n", answer))
						m.viewport.SetContent(m.output.String())
						m.viewport.GotoBottom()

						return m, m.waitForProgress()
					}
				}
			case tea.KeyEsc:
				// Cancel and go back
				m.cancel()
				m.running = false
				return m, nil
			}

			// Update input field
			var cmd tea.Cmd
			m.inputField, cmd = m.inputField.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Normal key handling when not waiting for input
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			m.cancel()
			m.running = false
			return m, nil
		}
	}

	// Update viewport
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m *WorkflowRunModel) handleProgress(msg ExecutionProgressMsg) (tea.Model, tea.Cmd) {
	progress := msg.ExecutionProgress

	// Handle completion
	if progress.Done {
		m.running = false
		m.completed = true
		if progress.Status == "error" {
			m.errMsg = progress.Output
		}
		return m, nil
	}

	// Update node status
	for i := range m.nodes {
		if m.nodes[i].ID == progress.NodeID {
			switch progress.Status {
			case "started":
				m.nodes[i].Status = NodeRunning
				m.currentNode = i
			case "completed":
				m.nodes[i].Status = NodeCompleted
			case "error":
				m.nodes[i].Status = NodeError
				m.nodes[i].Output = progress.Output
			case "waiting_input":
				m.nodes[i].Status = NodeWaitingInput
				m.waitingInput = true
				m.questionText = progress.Question
				m.options = progress.Options
				m.inputField.Focus()
			case "output":
				m.nodes[i].Output += progress.Output
				m.output.WriteString(progress.Output)
				m.viewport.SetContent(m.output.String())
				m.viewport.GotoBottom()
			}
			break
		}
	}

	// Continue listening unless waiting for input
	if m.waitingInput {
		return m, nil
	}

	return m, m.waitForProgress()
}

func (m *WorkflowRunModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1)

	workflowName := m.workflow.Name
	if workflowName == "" {
		workflowName = "Workflow"
	}

	statusText := ""
	if m.running {
		elapsed := time.Since(m.startTime)
		statusStyle := lipgloss.NewStyle().Foreground(accentColor)
		statusText = statusStyle.Render(fmt.Sprintf(" %s Running... (%ds)", m.spinner.View(), int(elapsed.Seconds())))
	} else if m.completed {
		if m.errMsg != "" {
			statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(" [Error]")
		} else {
			statusText = lipgloss.NewStyle().Foreground(secondaryColor).Render(" [Completed]")
		}
	}

	header := headerStyle.Render("Workflow: "+workflowName) + statusText

	// Node list (left panel)
	nodeListStyle := lipgloss.NewStyle().
		Border(getBorder()).
		BorderForeground(primaryColor).
		Padding(0, 1).
		Width(28)

	var nodeList strings.Builder
	nodeList.WriteString(mutedStyle.Render("Nodes") + "\n")
	nodeList.WriteString(strings.Repeat("-", 24) + "\n")

	for i, node := range m.nodes {
		icon := m.getStatusIcon(node.Status, i == m.currentNode && m.running)
		name := node.Name
		if len(name) > 18 {
			name = name[:15] + "..."
		}

		style := normalStyle
		if node.Status == NodeRunning {
			style = selectedStyle
		} else if node.Status == NodeCompleted {
			style = lipgloss.NewStyle().Foreground(secondaryColor)
		} else if node.Status == NodeError {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
		} else if node.Status == NodeWaitingInput {
			style = lipgloss.NewStyle().Foreground(accentColor)
		}

		nodeList.WriteString(fmt.Sprintf("%s %s\n", icon, style.Render(name)))
	}

	// Output panel (right panel)
	outputStyle := lipgloss.NewStyle().
		Border(getBorder()).
		BorderForeground(secondaryColor).
		Padding(0, 1).
		Width(m.width - 34)

	var outputContent strings.Builder
	outputContent.WriteString(mutedStyle.Render("Output") + "\n")
	outputContent.WriteString(strings.Repeat("-", m.width-40) + "\n")

	// Show question if waiting for input
	if m.waitingInput {
		questionStyle := lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)
		outputContent.WriteString("\n")
		outputContent.WriteString(questionStyle.Render("Question: "+m.questionText) + "\n")

		if len(m.options) > 0 {
			outputContent.WriteString(mutedStyle.Render("Options: "))
			for i, opt := range m.options {
				outputContent.WriteString(fmt.Sprintf("[%d] %s ", i+1, opt))
			}
			outputContent.WriteString("\n")
		}
		outputContent.WriteString("\n")
		outputContent.WriteString(inputStyle.Render(m.inputField.View()))
		outputContent.WriteString("\n")
	} else {
		outputContent.WriteString(m.viewport.View())
	}

	// Combine panels
	leftPanel := nodeListStyle.Render(nodeList.String())
	rightPanel := outputStyle.Render(outputContent.String())

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)

	// Help text
	var helpText string
	if m.waitingInput {
		helpText = helpStyle.Render("Enter: submit answer | Esc: cancel")
	} else if m.running {
		helpText = helpStyle.Render(fmt.Sprintf("Running node %d/%d | Esc: cancel", m.currentNode+1, len(m.nodes)))
	} else {
		helpText = helpStyle.Render("Esc: back to workflows")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		mainContent,
		"",
		helpText,
	)
}

func (m *WorkflowRunModel) getStatusIcon(status NodeStatus, isCurrentRunning bool) string {
	switch status {
	case NodePending:
		return "[ ]"
	case NodeRunning:
		if isCurrentRunning {
			return "[" + m.spinner.View() + "]"
		}
		return "[>]"
	case NodeCompleted:
		return "[*]"
	case NodeError:
		return "[!]"
	case NodeWaitingInput:
		return "[?]"
	default:
		return "[ ]"
	}
}

// IsRunning returns true if the workflow is still executing
func (m *WorkflowRunModel) IsRunning() bool {
	return m.running
}

// IsCompleted returns true if the workflow has finished
func (m *WorkflowRunModel) IsCompleted() bool {
	return m.completed
}

// Cancel stops the workflow execution
func (m *WorkflowRunModel) Cancel() {
	m.cancel()
	m.running = false
}
