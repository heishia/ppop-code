package agents

import (
	"context"
	"fmt"
)

type AgentType string

const (
	AgentTypeClaude AgentType = "claude"
	AgentTypeOpenAI AgentType = "openai"
	AgentTypeGemini AgentType = "gemini"
)

type AgentConfig struct {
	Name     string
	Type     AgentType
	Model    string
	APIKey   string
	BaseURL  string
	MaxTokens int
}

type Response struct {
	Content    string
	Model      string
	TokensUsed int
	Error      error
}

type Agent interface {
	Execute(ctx context.Context, prompt string) (*Response, error)
	Status() string
	Name() string
	Model() string
}

func NewAgent(config AgentConfig) (Agent, error) {
	switch config.Type {
	case AgentTypeClaude:
		return NewClaudeAgent(config)
	case AgentTypeOpenAI:
		return NewOpenAIAgent(config)
	case AgentTypeGemini:
		return NewGeminiAgent(config)
	default:
		return nil, fmt.Errorf("unknown agent type: %s", config.Type)
	}
}

type BaseAgent struct {
	config AgentConfig
	status string
}

func (a *BaseAgent) Name() string {
	return a.config.Name
}

func (a *BaseAgent) Model() string {
	return a.config.Model
}

func (a *BaseAgent) Status() string {
	return a.status
}

func (a *BaseAgent) SetStatus(status string) {
	a.status = status
}
