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
	// Get config path
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".ppopcode", "config.yaml")

	// Load configuration (will use defaults if file doesn't exist)
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Initialize session manager
	historyDir := filepath.Join(homeDir, cfg.Session.HistoryDir)
	sess := session.NewManager(historyDir, cfg.Session.MaxHistory)

	// Initialize orchestrator with agent configs
	agentConfigs := cfg.ToAgentConfigs()
	orch := orchestrator.New(agentConfigs)

	// Create app with dependencies
	app := tui.NewAppWithDeps(orch, sess, cfg)

	// Create program with alt screen (full terminal takeover)
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running ppopcode: %v\n", err)
		os.Exit(1)
	}
}
