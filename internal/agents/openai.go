package agents

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIAgent struct {
	BaseAgent
	client *openai.Client
}

func NewOpenAIAgent(config AgentConfig) (*OpenAIAgent, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	if apiKey == "" {
		return &OpenAIAgent{
			BaseAgent: BaseAgent{
				config: config,
				status: "no_api_key",
			},
		}, nil
	}

	clientConfig := openai.DefaultConfig(apiKey)
	if config.BaseURL != "" {
		clientConfig.BaseURL = config.BaseURL
	}

	client := openai.NewClientWithConfig(clientConfig)

	return &OpenAIAgent{
		BaseAgent: BaseAgent{
			config: config,
			status: "ready",
		},
		client: client,
	}, nil
}

func (a *OpenAIAgent) Execute(ctx context.Context, prompt string) (*Response, error) {
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

	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: a.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens: maxTokens,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("openai api error: %w", err)
	}

	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return &Response{
		Content:    content,
		Model:      a.config.Model,
		TokensUsed: resp.Usage.TotalTokens,
	}, nil
}
