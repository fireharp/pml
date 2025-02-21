package llm

import (
	"context"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Save current API key and restore it after test
	originalKey := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalKey)

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "No API key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "Valid API key",
			apiKey:  "test-key",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("OPENAI_API_KEY", tt.apiKey)
			client, err := NewClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestClientAsk(t *testing.T) {
	// Skip if no API key is set
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping test: no OpenAI API key set")
	}

	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	prompt := "What is 2+2?"
	
	response, err := client.Ask(ctx, prompt)
	if err != nil {
		t.Errorf("Ask() error = %v", err)
		return
	}
	if response == "" {
		t.Error("Ask() returned empty response")
	}
} 