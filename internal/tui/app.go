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
	ViewHowTo
	ViewSetup
	ViewChat
	ViewWorkflow
	ViewWorkflowRun
	ViewSettings
	ViewAbout
)

type KeyMap struct {
	Quit  key.Binding
	Back  key.Binding
	Enter key.Binding
	Up    key.Binding
	Down  key.Binding
	Help  key.Binding
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
	howto        *HowToModel
	setup        *SetupModel
	chat         *ChatModel
	workflow     *WorkflowModel
	workflowRun  *WorkflowRunModel
	settings     *SettingsModel
	about        *AboutModel
	width        int
	height       int
	keys         KeyMap
	orchestrator *orchestrator.Orchestrator
	session      *session.Manager
	config       *config.Config
	ready        bool // true after first WindowSizeMsg is received
}

func NewApp() *App {
	return &App{
		currentView: ViewMenu,
		menu:        NewMenuModel(),
		howto:       NewHowToModel(),
		setup:       NewSetupModel(),
		chat:        NewChatModel(),
		workflow:    NewWorkflowModel(),
		settings:    NewSettingsModel(),
		about:       NewAboutModel(),
		keys:        DefaultKeyMap,
	}
}

func NewAppWithDeps(orch *orchestrator.Orchestrator, sess *session.Manager, cfg *config.Config) *App {
	chat := NewChatModelWithOrchestrator(orch, sess)
	return &App{
		currentView:  ViewMenu,
		menu:         NewMenuModel(),
		howto:        NewHowToModel(),
		setup:        NewSetupModel(),
		chat:         chat,
		workflow:     NewWorkflowModel(),
		settings:     NewSettingsModelWithConfig(cfg),
		about:        NewAboutModel(),
		keys:         DefaultKeyMap,
		orchestrator: orch,
		session:      sess,
		config:       cfg,
	}
}

func (a *App) Init() tea.Cmd {
	// Return nil - tea.WithAltScreen in main.go handles screen setup
	// The window size message will be sent automatically
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.howto.SetSize(msg.Width, msg.Height-4)
		a.setup.SetSize(msg.Width, msg.Height-4)
		a.chat.SetSize(msg.Width, msg.Height-4)
		a.workflow.SetSize(msg.Width, msg.Height-4)
		if a.workflowRun != nil {
			a.workflowRun.SetSize(msg.Width, msg.Height-4)
		}
		a.settings.SetSize(msg.Width, msg.Height-4)
		a.about.SetSize(msg.Width, msg.Height-4)
		return a, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Quit):
			if a.currentView == ViewMenu {
				return a, tea.Quit
			}
			// Check if current view has sub-view first
			if a.currentView == ViewWorkflow && a.workflow.HasSubView() {
				a.workflow.CloseSubView()
				return a, nil
			}
			// Handle workflow run - cancel if running
			if a.currentView == ViewWorkflowRun && a.workflowRun != nil {
				if a.workflowRun.IsRunning() {
					a.workflowRun.Cancel()
				}
				a.currentView = ViewWorkflow
				return a, nil
			}
			a.currentView = ViewMenu
			a.menu.Selected = -1 // Reset menu selection
			return a, nil

		case key.Matches(msg, a.keys.Back):
			if a.currentView != ViewMenu {
				// Check if current view has sub-view first
				if a.currentView == ViewWorkflow && a.workflow.HasSubView() {
					a.workflow.CloseSubView()
					return a, nil
				}
				// Handle workflow run - go back to workflow list
				if a.currentView == ViewWorkflowRun && a.workflowRun != nil {
					if a.workflowRun.IsRunning() {
						a.workflowRun.Cancel()
					}
					a.currentView = ViewWorkflow
					return a, nil
				}
				a.currentView = ViewMenu
				a.menu.Selected = -1 // Reset menu selection
				return a, nil
			}
		}

	case WorkflowLoadedMsg:
		// Handle workflow loaded - create run model and switch view
		if msg.Error != nil {
			// TODO: Show error in workflow view
			return a, nil
		}
		a.workflowRun = NewWorkflowRunModel(msg.Workflow, a.orchestrator)
		a.workflowRun.SetSize(a.width, a.height-4)
		a.currentView = ViewWorkflowRun
		return a, a.workflowRun.Init()

	case ExecutionProgressMsg:
		// Forward execution progress to workflow run
		if a.currentView == ViewWorkflowRun && a.workflowRun != nil {
			newWfRun, wfRunCmd := a.workflowRun.Update(msg)
			a.workflowRun = newWfRun.(*WorkflowRunModel)
			return a, wfRunCmd
		}

	case StreamUpdateMsg:
		// Forward streaming updates directly to chat
		if a.currentView == ViewChat {
			newChat, chatCmd := a.chat.Update(msg)
			a.chat = newChat.(*ChatModel)
			return a, chatCmd
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
			case 0: // How to Start
				a.currentView = ViewHowTo
			case 1: // Link Accounts
				a.currentView = ViewSetup
				return a, a.setup.checkStatus
			case 2: // Start with Chat
				a.currentView = ViewChat
				a.chat.Focus()
			case 3: // Start with Workflow
				a.currentView = ViewWorkflow
				a.workflow.Reset() // Reset state when entering
			case 4: // Settings
				a.currentView = ViewSettings
			case 5: // About
				a.currentView = ViewAbout
			}
			a.menu.Selected = -1
		}

	case ViewHowTo:
		newHowTo, howtoCmd := a.howto.Update(msg)
		a.howto = newHowTo.(*HowToModel)
		cmd = howtoCmd

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

	case ViewWorkflowRun:
		if a.workflowRun != nil {
			newWfRun, wfRunCmd := a.workflowRun.Update(msg)
			a.workflowRun = newWfRun.(*WorkflowRunModel)
			cmd = wfRunCmd

			// If workflow run completed and user pressed back, return to workflow list
			if !a.workflowRun.IsRunning() && a.workflowRun.IsCompleted() {
				// Stay in workflow run view to show results
			}
		}

	case ViewSettings:
		newSettings, settingsCmd := a.settings.Update(msg)
		a.settings = newSettings.(*SettingsModel)
		cmd = settingsCmd

	case ViewAbout:
		newAbout, aboutCmd := a.about.Update(msg)
		a.about = newAbout.(*AboutModel)
		cmd = aboutCmd
	}

	return a, cmd
}

func (a *App) View() string {
	switch a.currentView {
	case ViewMenu:
		return a.menu.View()
	case ViewHowTo:
		return a.howto.View()
	case ViewSetup:
		return a.setup.View()
	case ViewChat:
		return a.chat.View()
	case ViewWorkflow:
		return a.workflow.View()
	case ViewWorkflowRun:
		if a.workflowRun != nil {
			return a.workflowRun.View()
		}
		return a.workflow.View()
	case ViewSettings:
		return a.settings.View()
	case ViewAbout:
		return a.about.View()
	default:
		return a.menu.View()
	}
}
