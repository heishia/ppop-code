package agents

import (
	"context"
	"errors"
	"fmt"
	"io"
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

// ExecuteStream executes the prompt and streams output in real-time
func (a *OpenAIAgent) ExecuteStream(ctx context.Context, prompt string, stream chan<- StreamChunk) (*Response, error) {
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

	stream <- StreamChunk{Content: "Connecting to OpenAI...", Type: "status"}

	maxTokens := a.config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	req := openai.ChatCompletionRequest{
		Model: a.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens: maxTokens,
		Stream:    true,
	}

	streamer, err := a.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		stream <- StreamChunk{Content: fmt.Sprintf("Error: %v", err), Type: "error", Done: true}
		return nil, fmt.Errorf("openai stream error: %w", err)
	}
	defer streamer.Close()

	var fullContent string
	for {
		response, err := streamer.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			stream <- StreamChunk{Content: fmt.Sprintf("Stream error: %v", err), Type: "error", Done: true}
			return nil, fmt.Errorf("openai stream recv error: %w", err)
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta.Content
			if delta != "" {
				fullContent += delta
				stream <- StreamChunk{Content: delta, Type: "output"}
			}
		}
	}

	stream <- StreamChunk{Content: "Done", Type: "status", Done: true}

	return &Response{
		Content: fullContent,
		Model:   a.config.Model,
	}, nil
}
