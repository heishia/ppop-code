package config

import (
	"os"
	"path/filepath"

	"github.com/ppopcode/ppopcode/internal/agents"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App     AppConfig              `yaml:"app"`
	Agents  map[string]AgentConfig `yaml:"agents"`
	Cursor  CursorConfig           `yaml:"cursor"`
	Session SessionConfig          `yaml:"session"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Debug   bool   `yaml:"debug"`
}

type AgentConfig struct {
	Type      string `yaml:"type"`
	Model     string `yaml:"model"`
	APIKey    string `yaml:"api_key,omitempty"`
	BaseURL   string `yaml:"base_url,omitempty"`
	MaxTokens int    `yaml:"max_tokens,omitempty"`
	Role      string `yaml:"role"`
}

type CursorConfig struct {
	Command  string `yaml:"command"`
	Timeout  int    `yaml:"timeout"`
	MaxRetry int    `yaml:"max_retry"`
}

type SessionConfig struct {
	SaveHistory bool   `yaml:"save_history"`
	HistoryDir  string `yaml:"history_dir"`
	MaxHistory  int    `yaml:"max_history"`
}

func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:    "ppopcode",
			Version: "1.0.0",
			Debug:   false,
		},
		Agents: map[string]AgentConfig{
			"orchestrator": {
				Type:      "claude",
				Model:     "claude-opus-4-20250514",
				Role:      "Main orchestrator - analyzes tasks and routes to appropriate agents",
				MaxTokens: 4096,
			},
			"gemini": {
				Type:      "gemini",
				Model:     "gemini-2.0-flash",
				Role:      "Frontend specialist - UX/UI development",
				MaxTokens: 4096,
			},
			"gpt": {
				Type:      "openai",
				Model:     "gpt-4o",
				Role:      "Design & debugging specialist",
				MaxTokens: 4096,
			},
			"sonnet": {
				Type:      "claude",
				Model:     "claude-sonnet-4-20250514",
				Role:      "General coding tasks",
				MaxTokens: 4096,
			},
		},
		Cursor: CursorConfig{
			Command:  "cursor-agent",
			Timeout:  300,
			MaxRetry: 2,
		},
		Session: SessionConfig{
			SaveHistory: true,
			HistoryDir:  ".ppopcode/history",
			MaxHistory:  100,
		},
	}
}

func Load(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) ToAgentConfigs() map[string]agents.AgentConfig {
	configs := make(map[string]agents.AgentConfig)

	for name, ac := range c.Agents {
		var agentType agents.AgentType
		switch ac.Type {
		case "claude":
			agentType = agents.AgentTypeClaude
		case "openai":
			agentType = agents.AgentTypeOpenAI
		case "gemini":
			agentType = agents.AgentTypeGemini
		}

		configs[name] = agents.AgentConfig{
			Name:      name,
			Type:      agentType,
			Model:     ac.Model,
			APIKey:    ac.APIKey,
			BaseURL:   ac.BaseURL,
			MaxTokens: ac.MaxTokens,
		}
	}

	return configs
}
