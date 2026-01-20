package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ppopcode/ppopcode/internal/config"
	"github.com/ppopcode/ppopcode/internal/orchestrator"
	"github.com/ppopcode/ppopcode/internal/session"
)

type ViewState int

const (
	ViewMenu ViewState = iota
	ViewSetup
	ViewChat
	ViewWorkflow
	ViewSettings
)

type KeyMap struct {
	Quit   key.Binding
	Back   key.Binding
	Enter  key.Binding
	Up     key.Binding
	Down   key.Binding
	Help   key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

type App struct {
	currentView  ViewState
	menu         *MenuModel
	setup        *SetupModel
	chat         *ChatModel
	workflow     *WorkflowModel
	width        int
	height       int
	keys         KeyMap
	orchestrator *orchestrator.Orchestrator
	session      *session.Manager
	config       *config.Config
}

func NewApp() *App {
	return &App{
		currentView: ViewMenu,
		menu:        NewMenuModel(),
		setup:       NewSetupModel(),
		chat:        NewChatModel(),
		workflow:    NewWorkflowModel(),
		keys:        DefaultKeyMap,
	}
}

func NewAppWithDeps(orch *orchestrator.Orchestrator, sess *session.Manager, cfg *config.Config) *App {
	chat := NewChatModelWithOrchestrator(orch, sess)
	return &App{
		currentView:  ViewMenu,
		menu:         NewMenuModel(),
		setup:        NewSetupModel(),
		chat:         chat,
		workflow:     NewWorkflowModel(),
		keys:         DefaultKeyMap,
		orchestrator: orch,
		session:      sess,
		config:       cfg,
	}
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.setup.SetSize(msg.Width, msg.Height-4)
		a.chat.SetSize(msg.Width, msg.Height-4)
		a.workflow.SetSize(msg.Width, msg.Height-4)
		return a, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Quit):
			if a.currentView == ViewMenu {
				return a, tea.Quit
			}
			a.currentView = ViewMenu
			return a, nil

		case key.Matches(msg, a.keys.Back):
			if a.currentView != ViewMenu {
				a.currentView = ViewMenu
				return a, nil
			}
		}
	}

	var cmd tea.Cmd
	switch a.currentView {
	case ViewMenu:
		newMenu, menuCmd := a.menu.Update(msg)
		a.menu = newMenu.(*MenuModel)
		cmd = menuCmd

		if a.menu.Selected != -1 {
			switch a.menu.Selected {
			case 0:
				a.currentView = ViewSetup
				return a, a.setup.checkStatus
			case 1:
				a.currentView = ViewChat
				a.chat.Focus()
			case 2:
				a.currentView = ViewWorkflow
			}
			a.menu.Selected = -1
		}

	case ViewSetup:
		newSetup, setupCmd := a.setup.Update(msg)
		a.setup = newSetup.(*SetupModel)
		cmd = setupCmd

	case ViewChat:
		newChat, chatCmd := a.chat.Update(msg)
		a.chat = newChat.(*ChatModel)
		cmd = chatCmd

	case ViewWorkflow:
		newWorkflow, wfCmd := a.workflow.Update(msg)
		a.workflow = newWorkflow.(*WorkflowModel)
		cmd = wfCmd
	}

	return a, cmd
}

func (a *App) View() string {
	switch a.currentView {
	case ViewMenu:
		return a.menu.View()
	case ViewSetup:
		return a.setup.View()
	case ViewChat:
		return a.chat.View()
	case ViewWorkflow:
		return a.workflow.View()
	default:
		return a.menu.View()
	}
}
