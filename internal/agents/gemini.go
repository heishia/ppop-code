package agents

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type GeminiAgent struct {
	BaseAgent
	client *genai.Client
}

func NewGeminiAgent(config AgentConfig) (*GeminiAgent, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}

	if apiKey == "" {
		return &GeminiAgent{
			BaseAgent: BaseAgent{
				config: config,
				status: "no_api_key",
			},
		}, nil
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &GeminiAgent{
		BaseAgent: BaseAgent{
			config: config,
			status: "ready",
		},
		client: client,
	}, nil
}

func (a *GeminiAgent) Execute(ctx context.Context, prompt string) (*Response, error) {
	if a.client == nil {
		return &Response{
			Content: fmt.Sprintf("[%s - %s] API key not configured. This is a placeholder response.\n\nYour request: %s", a.config.Name, a.config.Model, prompt),
			Model:   a.config.Model,
		}, nil
	}

	a.SetStatus("processing")
	defer a.SetStatus("ready")

	result, err := a.client.Models.GenerateContent(ctx, a.config.Model, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("gemini api error: %w", err)
	}

	content := ""
	if result != nil && len(result.Candidates) > 0 {
		candidate := result.Candidates[0]
		if candidate.Content != nil && len(candidate.Content.Parts) > 0 {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					content += part.Text
				}
			}
		}
	}

	return &Response{
		Content: content,
		Model:   a.config.Model,
	}, nil
}
