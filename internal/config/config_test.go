package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ppopcode/ppopcode/internal/agents"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() should not return nil")
	}

	// Check app config
	if config.App.Name != "ppopcode" {
		t.Errorf("App.Name = %q, want %q", config.App.Name, "ppopcode")
	}

	if config.App.Version != "1.0.0" {
		t.Errorf("App.Version = %q, want %q", config.App.Version, "1.0.0")
	}

	// Check agents are defined
	expectedAgents := []string{"orchestrator", "gemini", "gpt", "sonnet"}
	for _, name := range expectedAgents {
		if _, exists := config.Agents[name]; !exists {
			t.Errorf("Missing agent: %s", name)
		}
	}

	// Check cursor config
	if config.Cursor.Command != "cursor-agent" {
		t.Errorf("Cursor.Command = %q, want %q", config.Cursor.Command, "cursor-agent")
	}

	// Check session config
	if !config.Session.SaveHistory {
		t.Error("Session.SaveHistory should be true by default")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	config, err := Load("/nonexistent/path/config.yaml")

	if err != nil {
		t.Fatalf("Load() should not error for nonexistent file: %v", err)
	}

	if config == nil {
		t.Fatal("Load() should return default config for nonexistent file")
	}

	// Should have default values
	if config.App.Name != "ppopcode" {
		t.Errorf("Should have default app name")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create custom config
	config := DefaultConfig()
	config.App.Name = "custom-name"
	config.App.Debug = true

	// Save
	if err := config.Save(configPath); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.App.Name != "custom-name" {
		t.Errorf("loaded.App.Name = %q, want %q", loaded.App.Name, "custom-name")
	}

	if !loaded.App.Debug {
		t.Error("loaded.App.Debug should be true")
	}
}

func TestToAgentConfigs(t *testing.T) {
	config := DefaultConfig()

	agentConfigs := config.ToAgentConfigs()

	if agentConfigs == nil {
		t.Fatal("ToAgentConfigs() should not return nil")
	}

	// Check sonnet agent
	sonnet, exists := agentConfigs["sonnet"]
	if !exists {
		t.Fatal("sonnet agent config should exist")
	}

	if sonnet.Type != agents.AgentTypeClaude {
		t.Errorf("sonnet.Type = %v, want %v", sonnet.Type, agents.AgentTypeClaude)
	}

	if sonnet.Name != "sonnet" {
		t.Errorf("sonnet.Name = %q, want %q", sonnet.Name, "sonnet")
	}

	// Check gemini agent
	gemini, exists := agentConfigs["gemini"]
	if !exists {
		t.Fatal("gemini agent config should exist")
	}

	if gemini.Type != agents.AgentTypeGemini {
		t.Errorf("gemini.Type = %v, want %v", gemini.Type, agents.AgentTypeGemini)
	}

	// Check gpt agent
	gpt, exists := agentConfigs["gpt"]
	if !exists {
		t.Fatal("gpt agent config should exist")
	}

	if gpt.Type != agents.AgentTypeOpenAI {
		t.Errorf("gpt.Type = %v, want %v", gpt.Type, agents.AgentTypeOpenAI)
	}
}

func TestAgentConfigFields(t *testing.T) {
	config := DefaultConfig()

	// Test orchestrator config
	orch := config.Agents["orchestrator"]
	if orch.Type != "claude" {
		t.Errorf("orchestrator.Type = %q, want %q", orch.Type, "claude")
	}

	if orch.MaxTokens != 4096 {
		t.Errorf("orchestrator.MaxTokens = %d, want %d", orch.MaxTokens, 4096)
	}

	if orch.Role == "" {
		t.Error("orchestrator.Role should not be empty")
	}
}

func TestCursorConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Cursor.Timeout != 300 {
		t.Errorf("Cursor.Timeout = %d, want %d", config.Cursor.Timeout, 300)
	}

	if config.Cursor.MaxRetry != 2 {
		t.Errorf("Cursor.MaxRetry = %d, want %d", config.Cursor.MaxRetry, 2)
	}
}

func TestSessionConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Session.HistoryDir != ".ppopcode/history" {
		t.Errorf("Session.HistoryDir = %q, want %q", config.Session.HistoryDir, ".ppopcode/history")
	}

	if config.Session.MaxHistory != 100 {
		t.Errorf("Session.MaxHistory = %d, want %d", config.Session.MaxHistory, 100)
	}
}
