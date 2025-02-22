package parser

import (
	"context"
	"time"
)

// mockLLM implements LLMClient for testing
type mockLLM struct {
	response string
	err      error
	callback func()
	Delay    time.Duration // configurable delay for Ask
}

func (m *mockLLM) Ask(ctx context.Context, prompt string) (string, error) {
	if m.callback != nil {
		m.callback()
	}
	// Use m.Delay if provided; otherwise, default to 300ms.
	totalDelay := m.Delay
	if totalDelay == 0 {
		totalDelay = 300 * time.Millisecond  // Longer default delay for cancellation test
	}
	interval := 10 * time.Millisecond
	elapsed := time.Duration(0)
	for elapsed < totalDelay {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			time.Sleep(interval)
			elapsed += interval
		}
	}
	return m.response, m.err
}

func (m *mockLLM) Summarize(ctx context.Context, text string) (string, error) {
	if m.callback != nil {
		m.callback()
	}
	return "Summary: " + text, m.err
}
