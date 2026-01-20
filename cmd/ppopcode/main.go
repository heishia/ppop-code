package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ppopcode/ppopcode/internal/config"
	"github.com/ppopcode/ppopcode/internal/orchestrator"
	"github.com/ppopcode/ppopcode/internal/session"
	"github.com/ppopcode/ppopcode/internal/tui"
)

func main() {
	// Load configuration
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configPath := filepath.Join(homeDir, ".ppopcode", "config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		// Use default config if load fails
		cfg = config.DefaultConfig()
	}

	// Initialize session manager
	historyDir := cfg.Session.HistoryDir
	if !filepath.IsAbs(historyDir) {
		historyDir = filepath.Join(homeDir, ".ppopcode", historyDir)
	}
	sessionMgr := session.NewManager(historyDir, cfg.Session.MaxHistory)

	// Initialize orchestrator with agent configs
	agentConfigs := cfg.ToAgentConfigs()
	orch := orchestrator.New(agentConfigs)

	// Create and run TUI app
	app := tui.NewAppWithDeps(orch, sessionMgr, cfg)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
