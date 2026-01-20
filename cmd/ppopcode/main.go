package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ppopcode/ppopcode/internal/config"
	"github.com/ppopcode/ppopcode/internal/orchestrator"
	"github.com/ppopcode/ppopcode/internal/session"
	"github.com/ppopcode/ppopcode/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, err := config.Load("config/ppopcode.yaml")
	if err != nil {
		fmt.Printf("Warning: Failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	agentConfigs := cfg.ToAgentConfigs()
	orch := orchestrator.New(agentConfigs)

	historyDir := cfg.Session.HistoryDir
	if !filepath.IsAbs(historyDir) {
		cwd, _ := os.Getwd()
		historyDir = filepath.Join(cwd, historyDir)
	}
	sessionMgr := session.NewManager(historyDir, cfg.Session.MaxHistory)

	app := tui.NewAppWithDeps(orch, sessionMgr, cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running ppopcode: %v\n", err)
		os.Exit(1)
	}
}
