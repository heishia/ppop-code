package agents

import (
	"testing"
)

func TestAgentTypes(t *testing.T) {
	types := []AgentType{
		AgentTypeClaude,
	}

	for _, agentType := range types {
		if agentType == "" {
			t.Error("AgentType should not be empty")
		}
	}

	// Verify expected values
	if AgentTypeClaude != "claude" {
		t.Errorf("AgentTypeClaude = %q, want %q", AgentTypeClaude, "claude")
	}
}

func TestNewAgentUnknownType(t *testing.T) {
	config := AgentConfig{
		Name: "test",
		Type: AgentType("unknown"),
	}

	_, err := NewAgent(config)
	if err == nil {
		t.Error("NewAgent() should return error for unknown type")
	}
}

func TestNewAgentClaude(t *testing.T) {
	config := AgentConfig{
		Name:  "test-claude",
		Type:  AgentTypeClaude,
		Model: "claude-sonnet",
	}

	agent, err := NewAgent(config)
	if err != nil {
		t.Fatalf("NewAgent(claude) error: %v", err)
	}

	if agent == nil {
		t.Fatal("NewAgent(claude) should not return nil")
	}

	if agent.Name() != "test-claude" {
		t.Errorf("agent.Name() = %q, want %q", agent.Name(), "test-claude")
	}

	if agent.Model() != "claude-sonnet" {
		t.Errorf("agent.Model() = %q, want %q", agent.Model(), "claude-sonnet")
	}
}

func TestBaseAgent(t *testing.T) {
	base := &BaseAgent{
		config: AgentConfig{
			Name:  "base-test",
			Model: "test-model",
		},
		status: "ready",
	}

	if base.Name() != "base-test" {
		t.Errorf("Name() = %q, want %q", base.Name(), "base-test")
	}

	if base.Model() != "test-model" {
		t.Errorf("Model() = %q, want %q", base.Model(), "test-model")
	}

	if base.Status() != "ready" {
		t.Errorf("Status() = %q, want %q", base.Status(), "ready")
	}

	base.SetStatus("processing")
	if base.Status() != "processing" {
		t.Errorf("Status() after SetStatus = %q, want %q", base.Status(), "processing")
	}
}

func TestAgentConfig(t *testing.T) {
	config := AgentConfig{
		Name:      "test",
		Type:      AgentTypeClaude,
		Model:     "claude-sonnet",
		APIKey:    "test-key",
		BaseURL:   "https://api.example.com",
		MaxTokens: 4096,
	}

	if config.Name != "test" {
		t.Errorf("config.Name = %q, want %q", config.Name, "test")
	}

	if config.Type != AgentTypeClaude {
		t.Errorf("config.Type = %q, want %q", config.Type, AgentTypeClaude)
	}

	if config.Model != "claude-sonnet" {
		t.Errorf("config.Model = %q, want %q", config.Model, "claude-sonnet")
	}

	if config.APIKey != "test-key" {
		t.Errorf("config.APIKey = %q, want %q", config.APIKey, "test-key")
	}

	if config.BaseURL != "https://api.example.com" {
		t.Errorf("config.BaseURL = %q, want %q", config.BaseURL, "https://api.example.com")
	}

	if config.MaxTokens != 4096 {
		t.Errorf("config.MaxTokens = %d, want %d", config.MaxTokens, 4096)
	}
}

func TestResponse(t *testing.T) {
	resp := &Response{
		Content:    "Hello, world!",
		Model:      "claude-sonnet",
		TokensUsed: 100,
	}

	if resp.Content != "Hello, world!" {
		t.Errorf("resp.Content = %q, want %q", resp.Content, "Hello, world!")
	}

	if resp.Model != "claude-sonnet" {
		t.Errorf("resp.Model = %q, want %q", resp.Model, "claude-sonnet")
	}

	if resp.TokensUsed != 100 {
		t.Errorf("resp.TokensUsed = %d, want %d", resp.TokensUsed, 100)
	}
}

func TestStreamChunk(t *testing.T) {
	chunk := StreamChunk{
		Content: "streaming content",
		Type:    "output",
		Done:    false,
	}

	if chunk.Content != "streaming content" {
		t.Errorf("chunk.Content = %q, want %q", chunk.Content, "streaming content")
	}

	if chunk.Type != "output" {
		t.Errorf("chunk.Type = %q, want %q", chunk.Type, "output")
	}

	if chunk.Done {
		t.Error("chunk.Done should be false")
	}
}
