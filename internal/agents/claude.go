package agents

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeAgent struct {
	BaseAgent
	client *anthropic.Client
}

func NewClaudeAgent(config AgentConfig) (*ClaudeAgent, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	if apiKey == "" {
		return &ClaudeAgent{
			BaseAgent: BaseAgent{
				config: config,
				status: "no_api_key",
			},
		}, nil
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &ClaudeAgent{
		BaseAgent: BaseAgent{
			config: config,
			status: "ready",
		},
		client: client,
	}, nil
}

func (a *ClaudeAgent) Execute(ctx context.Context, prompt string) (*Response, error) {
	if a.client == nil {
		return &Response{
			Content: fmt.Sprintf("[%s - %s] API key not configured. This is a placeholder response.\n\nYour request: %s", a.config.Name, a.config.Model, prompt),
			Model:   a.config.Model,
		}, nil
	}

	a.SetStatus("processing")
	defer a.SetStatus("ready")

	maxTokens := a.config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(a.config.Model),
		MaxTokens: anthropic.F(int64(maxTokens)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("claude api error: %w", err)
	}

	content := ""
	for _, block := range message.Content {
		if block.Type == anthropic.ContentBlockTypeText {
			content += block.Text
		}
	}

	return &Response{
		Content:    content,
		Model:      a.config.Model,
		TokensUsed: int(message.Usage.InputTokens + message.Usage.OutputTokens),
	}, nil
}
