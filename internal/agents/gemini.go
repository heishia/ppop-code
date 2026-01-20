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

// ExecuteStream executes the prompt and streams output in real-time
func (a *GeminiAgent) ExecuteStream(ctx context.Context, prompt string, stream chan<- StreamChunk) (*Response, error) {
	defer close(stream)

	if a.client == nil {
		stream <- StreamChunk{Content: "API key not configured", Type: "error", Done: true}
		return &Response{
			Content: fmt.Sprintf("[%s] API key not configured.", a.config.Name),
			Model:   a.config.Model,
		}, nil
	}

	a.SetStatus("processing")
	defer a.SetStatus("ready")

	stream <- StreamChunk{Content: "Connecting to Gemini...", Type: "status"}

	// Use streaming API - iter.Seq2 returns (response, error) pairs
	iter := a.client.Models.GenerateContentStream(ctx, a.config.Model, genai.Text(prompt), nil)

	var fullContent string
	for resp, err := range iter {
		if err != nil {
			stream <- StreamChunk{Content: fmt.Sprintf("Stream error: %v", err), Type: "error", Done: true}
			return nil, fmt.Errorf("gemini stream error: %w", err)
		}

		if resp != nil && len(resp.Candidates) > 0 {
			candidate := resp.Candidates[0]
			if candidate.Content != nil && len(candidate.Content.Parts) > 0 {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						fullContent += part.Text
						stream <- StreamChunk{Content: part.Text, Type: "output"}
					}
				}
			}
		}
	}

	stream <- StreamChunk{Content: "Done", Type: "status", Done: true}

	return &Response{
		Content: fullContent,
		Model:   a.config.Model,
	}, nil
}
