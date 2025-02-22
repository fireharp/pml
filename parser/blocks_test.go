package parser

import (
	"strings"
	"testing"
)

// TestCalculateBlockChecksum verifies that checksums are consistent for the same content.
func TestCalculateBlockChecksum(t *testing.T) {
	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	blockA := Block{
		Type: ":ask",
		Content: []string{
			"What is 2+2?",
		},
	}
	blockB := Block{
		Type: ":ask",
		Content: []string{
			"   What is 2+2?   ", // same content but extra spaces
		},
	}

	checksumA := parser.calculateBlockChecksum(blockA)
	checksumB := parser.calculateBlockChecksum(blockB)
	if checksumA != checksumB {
		t.Errorf("Checksums differ for semantically identical blocks.\nA=%s\nB=%s\n", checksumA, checksumB)
	}
}

// TestParseBlocksWithEmptyLines verifies that empty lines don't cause extra blocks.
func TestParseBlocksWithEmptyLines(t *testing.T) {
	content := strings.Join([]string{
		":ask",
		"What is 2+2?",
		"",
		":--",
		"",
		":do",
		"",
		"Run some action",
		":--",
		"",
	}, "\n")

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	blocks, err := parser.parseBlocks(content)
	if err != nil {
		t.Fatalf("parseBlocks failed: %v", err)
	}

	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(blocks))
	}
}

// TestParseBlocksWithMultilineContent tests parsing blocks with multiple lines of content.
func TestParseBlocksWithMultilineContent(t *testing.T) {
	content := strings.Join([]string{
		":ask",
		"Line 1",
		"Line 2",
		"Line 3",
		":--",
	}, "\n")

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	blocks, err := parser.parseBlocks(content)
	if err != nil {
		t.Fatalf("parseBlocks failed: %v", err)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	if len(blocks[0].Content) != 3 {
		t.Errorf("Expected 3 lines in content, got %d", len(blocks[0].Content))
	}
}

// TestParseBlocksWithInvalidFormat tests handling of malformed blocks.
func TestParseBlocksWithInvalidFormat(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "Missing end marker",
			content: `:ask
What is 2+2?`,
			wantErr: true,
		},
		{
			name: "Empty block",
			content: `:ask
:--`,
			wantErr: false,
		},
		{
			name: "Invalid directive",
			content: `:invalid
Some content
:--`,
			wantErr: true,
		},
	}

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.parseBlocks(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBlocks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
