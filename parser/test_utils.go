package parser

import (
	"context"
	"time"
)

// mockLLM implements LLMClient for testing
type mockLLM struct {
	response string
	err      error
	callback func() // Add callback to track when Ask is called
}

func (m *mockLLM) Ask(ctx context.Context, prompt string) (string, error) {
	if m.callback != nil {
		m.callback()
	}
	time.Sleep(300 * time.Millisecond) // artificial delay to force cancellation
	return m.response, m.err
}

func (m *mockLLM) Summarize(ctx context.Context, text string) (string, error) {
	if m.callback != nil {
		m.callback()
	}
	return "Summary: " + text, m.err
}
