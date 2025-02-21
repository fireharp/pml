package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Client represents an LLM client
type Client struct {
	openaiClient *openai.Client
}

// NewClient creates a new LLM client
func NewClient() (*Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set. Please configure it in the PML extension settings")
	}

	return &Client{
		openaiClient: openai.NewClient(apiKey),
	}, nil
}

// Ask sends a prompt to the LLM and returns the response
func (c *Client) Ask(ctx context.Context, prompt string) (string, error) {
	resp, err := c.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4o-mini",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get LLM response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from LLM")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// Summarize generates a very short summary of the given text
func (c *Client) Summarize(ctx context.Context, text string) (string, error) {
	resp, err := c.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4o-mini",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: `You are a summarizer that creates extremely concise summaries. 
                             Keep summaries under 5 words. 
							 As short as possible. But not loosing the point.
							 For example:
							 "The capital of Japan is Tokyo." -> "Tokyo"
							 "Hello, world!" -> "Hello, world!"`,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Summarize this in under 5 words:\n" + text,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from LLM")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
