package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateUniqueResultName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-results-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")

	// Test basic name generation
	name1 := parser.generateUniqueResultName("mySourceFile.pml", 0, tmpDir)
	if !strings.HasPrefix(name1, "ask_") {
		t.Errorf("Expected name to start with 'ask_', got %s", name1)
	}

	// Test collision handling
	name2 := parser.generateUniqueResultName("mySourceFile.pml", 0, tmpDir)
	if name1 == name2 {
		t.Errorf("Expected unique names, but got collisions: %s == %s", name1, name2)
	}

	// Test different block indices
	name3 := parser.generateUniqueResultName("mySourceFile.pml", 1, tmpDir)
	if strings.HasPrefix(name3, name1) {
		t.Errorf("Names from different block indices should be different: %s vs %s", name1, name3)
	}
}

func TestFormatResult(t *testing.T) {
	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "Hello World",
			expected: `"Hello World"`,
		},
		{
			name:     "numeric value",
			input:    "123",
			expected: "123",
		},
		{
			name:     "string with quotes",
			input:    `Hello "World"`,
			expected: `"Hello \"World\""`,
		},
		{
			name:     "multiline string",
			input:    "Hello\nWorld",
			expected: `"Hello\nWorld"`,
		},
		{
			name:     "json-like string",
			input:    `{"key": "value"}`,
			expected: `"{\"key\": \"value\"}"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := parser.formatResult(tc.input)
			if got != tc.expected {
				t.Errorf("formatResult(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestWriteResult(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-results-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")

	block := Block{
		Type:    ":ask",
		Content: []string{"What is 2+2?"},
	}
	result := "4"
	resultFile := "test_result.pml"
	summary := "Test summary"

	err = parser.writeResult(block, result, resultFile, tmpDir, summary)
	if err != nil {
		t.Fatalf("writeResult failed: %v", err)
	}

	// Read the result file
	content, err := os.ReadFile(filepath.Join(tmpDir, resultFile))
	if err != nil {
		t.Fatal(err)
	}

	// Check content
	contentStr := string(content)
	if !strings.Contains(contentStr, summary) {
		t.Error("Result file missing summary")
	}
	if !strings.Contains(contentStr, block.Content[0]) {
		t.Error("Result file missing original question")
	}
	if !strings.Contains(contentStr, result) {
		t.Error("Result file missing result")
	}
}

func TestResultLinkGeneration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-results-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")

	// Create a test file and process it
	srcFile := filepath.Join(tmpDir, "test.pml")
	content := `:ask
What is 2+2?
:--`
	err = os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = parser.ProcessFile(nil, srcFile)
	if err != nil {
		t.Fatal(err)
	}

	// Read the processed file
	processedContent, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check for result link
	contentStr := string(processedContent)
	if !strings.Contains(contentStr, ":--(r/") {
		t.Error("Processed file missing result link")
	}

	// Extract result file name from link
	linkStart := strings.Index(contentStr, ":--(r/")
	linkEnd := strings.Index(contentStr[linkStart:], ")")
	if linkStart == -1 || linkEnd == -1 {
		t.Fatal("Could not find result link")
	}

	link := contentStr[linkStart : linkStart+linkEnd+1]
	resultFile := strings.Split(strings.Split(link, ":")[1], ")")[0]
	resultFile = strings.Trim(resultFile, `"`)

	// Verify result file exists
	if _, err := os.Stat(filepath.Join(tmpDir, resultFile)); os.IsNotExist(err) {
		t.Errorf("Result file %s does not exist", resultFile)
	}
}

func TestResultFileNaming(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-results-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")

	// Create multiple blocks in a file
	srcFile := filepath.Join(tmpDir, "multi.pml")
	content := `:ask
Q1
:--

:ask
Q2
:--

:ask
Q3
:--`
	err = os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = parser.ProcessFile(nil, srcFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check result files
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	resultFiles := make([]string, 0)
	for _, f := range files {
		if f.Name() != "multi.pml" && strings.HasSuffix(f.Name(), ".pml") {
			resultFiles = append(resultFiles, f.Name())
		}
	}

	if len(resultFiles) != 3 {
		t.Errorf("Expected 3 result files, got %d", len(resultFiles))
	}

	// Verify each result file has a unique name
	seen := make(map[string]bool)
	for _, f := range resultFiles {
		if seen[f] {
			t.Errorf("Duplicate result file name: %s", f)
		}
		seen[f] = true
	}
}
