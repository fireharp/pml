package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestProcessFileUnknownBlock tests that an unknown block directive returns an error.
func TestProcessFileUnknownBlock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-unknownblock-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "unknown.pml")
	content := `:foo
Some unhandled directive
:--
`
	err = os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	err = parser.ProcessFile(context.Background(), srcFile)
	if err == nil {
		t.Errorf("Expected error for unknown block directive, got nil")
	}
}

// TestProcessFileWithNestedBlocks tests handling of nested block structures
func TestProcessFileWithNestedBlocks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-nested-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := `:ask
What is this:
:ask
Nested question
:--
:--
`
	srcFile := filepath.Join(tmpDir, "nested.pml")
	err = os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	err = parser.ProcessFile(context.Background(), srcFile)
	if err == nil {
		t.Error("Expected error for nested blocks, got nil")
	}
}

// TestProcessFileWithMalformedBlocks tests handling of malformed block structures
func TestProcessFileWithMalformedBlocks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-malformed-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "missing end marker",
			content: `:ask
What is this?`,
			wantErr: true,
		},
		{
			name: "empty block",
			content: `:ask
:--`,
			wantErr: false,
		},
		{
			name: "multiple end markers",
			content: `:ask
What is this?
:--
:--`,
			wantErr: true,
		},
		{
			name: "no content between blocks",
			content: `:ask
:--
:ask
:--`,
			wantErr: false,
		},
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srcFile := filepath.Join(tmpDir, tc.name+".pml")
			err := os.WriteFile(srcFile, []byte(tc.content), 0644)
			if err != nil {
				t.Fatal(err)
			}

			err = parser.ProcessFile(context.Background(), srcFile)
			if (err != nil) != tc.wantErr {
				t.Errorf("ProcessFile() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestProcessFileWithComments tests handling of comments in PML files
func TestProcessFileWithComments(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-comments-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := `# This is a comment
:ask
What is 2+2?
# Another comment
:--
`
	srcFile := filepath.Join(tmpDir, "comments.pml")
	err = os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	err = parser.ProcessFile(context.Background(), srcFile)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	// Read processed file
	processedContent, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	// Comments should be preserved
	if !strings.Contains(string(processedContent), "# This is a comment") {
		t.Error("Comments were not preserved in processed file")
	}
}

// TestProcessFileWithWhitespace tests handling of various whitespace patterns
func TestProcessFileWithWhitespace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-whitespace-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := `

:ask

What is 2+2?

:--

:ask
What is 3+3?
:--

`
	srcFile := filepath.Join(tmpDir, "whitespace.pml")
	err = os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	err = parser.ProcessFile(context.Background(), srcFile)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	// Verify both blocks were processed
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	resultCount := 0
	for _, f := range files {
		if f.Name() != "whitespace.pml" && strings.HasSuffix(f.Name(), ".pml") {
			resultCount++
		}
	}

	if resultCount != 2 {
		t.Errorf("Expected 2 result files, got %d", resultCount)
	}
}

// TestProcessFileWithUTF8 tests handling of UTF-8 content
func TestProcessFileWithUTF8(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-utf8-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := `:ask
What is π?
:--

:ask
こんにちは世界
:--
`
	srcFile := filepath.Join(tmpDir, "utf8.pml")
	err = os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	err = parser.ProcessFile(context.Background(), srcFile)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	// Read result files and verify UTF-8 content is preserved
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	foundUTF8 := false
	for _, f := range files {
		if f.Name() != "utf8.pml" && strings.HasSuffix(f.Name(), ".pml") {
			content, err := os.ReadFile(filepath.Join(tmpDir, f.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(content), "π") || strings.Contains(string(content), "こんにちは世界") {
				foundUTF8 = true
				break
			}
		}
	}

	if !foundUTF8 {
		t.Error("UTF-8 content was not preserved in result files")
	}
}

// TestProcessFileWithLargeContent tests handling of large files
func TestProcessFileWithLargeContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-large-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a large file with many blocks
	var contentBuilder strings.Builder
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&contentBuilder, `:ask
Question %d
:--

`, i)
	}

	srcFile := filepath.Join(tmpDir, "large.pml")
	err = os.WriteFile(srcFile, []byte(contentBuilder.String()), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	err = parser.ProcessFile(context.Background(), srcFile)
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	// Verify all blocks were processed
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	resultCount := 0
	for _, f := range files {
		if f.Name() != "large.pml" && strings.HasSuffix(f.Name(), ".pml") {
			resultCount++
		}
	}

	if resultCount != 100 {
		t.Errorf("Expected 100 result files, got %d", resultCount)
	}
}
