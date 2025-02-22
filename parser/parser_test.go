package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
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
	return m.response, m.err
}

func (m *mockLLM) Summarize(ctx context.Context, text string) (string, error) {
	if m.callback != nil {
		m.callback()
	}
	return "Summary: " + text, m.err
}

func TestIsPMLFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Simple PML", "test.pml", true},
		{"Uppercase PML", "TEST.PML", true},
		{"Mixed case PML", "Test.PmL", true},
		{"Non PML", "test.txt", false},
		{"No extension", "test", false},
		{"Path with PML", "/path/to/test.pml", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPMLFile(tt.path); got != tt.expected {
				t.Errorf("IsPMLFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestParseBlocks(t *testing.T) {
	content := `:ask
What is 2+2?
:--
:do
Run some action
:--
:ask
Another question
:--
Result here
`
	mockLLM := &mockLLM{response: "Test response"}
	parser := NewParser(mockLLM, "sources", "compiled", "results")
	blocks, err := parser.parseBlocks(content)
	if err != nil {
		t.Fatal(err)
	}

	if len(blocks) != 3 {
		t.Errorf("Expected 3 blocks, got %d", len(blocks))
	}

	expectedTypes := []string{":ask", ":do", ":ask"}
	expectedContents := []string{"What is 2+2?", "Run some action", "Another question"}

	for i, block := range blocks {
		if block.Type != expectedTypes[i] {
			t.Errorf("Block %d: expected type %s, got %s", i, expectedTypes[i], block.Type)
		}
		if strings.TrimSpace(strings.Join(block.Content, "\n")) != expectedContents[i] {
			t.Errorf("Block %d: expected content %q, got %q", i, expectedContents[i], strings.Join(block.Content, "\n"))
		}
	}
}

func TestReplaceBlocksInContent(t *testing.T) {
	content := `:ask
What is 2+2?
:--

def some_code():
    print("hello")

:do
Run some action
:--

some_code()
`
	blocks := []Block{
		{Type: ":ask", Content: []string{"What is 2+2?"}},
		{Type: ":do", Content: []string{"Run some action"}},
	}

	mockLLM := &mockLLM{response: "Test response"}
	parser := NewParser(mockLLM, "sources", "compiled", "results")
	newContent := parser.replaceBlocksInContent(content, blocks)

	// Expected content should have imports at top and blocks replaced with Python equivalents
	expectedParts := []string{
		"# Auto-generated imports for PML blocks",
		"from src.pml.directives import process_ask, process_do",
		"",
		"# :ask",
		"result_0 = process_ask('''",
		"What is 2+2?",
		"''')",
		"# :--",
		"",
		"def some_code():",
		"    print(\"hello\")",
		"",
		"# :do",
		"result_1 = process_do('''",
		"Run some action",
		"''')",
		"# :--",
		"",
		"some_code()",
	}

	// Split both expected and actual content into lines for comparison
	expectedLines := strings.Join(expectedParts, "\n")
	// Normalize whitespace for comparison
	normalizeWhitespace := func(s string) string {
		lines := strings.Split(s, "\n")
		var normalized []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				normalized = append(normalized, strings.TrimSpace(line))
			}
		}
		return strings.Join(normalized, "\n")
	}

	if normalizeWhitespace(newContent) != normalizeWhitespace(expectedLines) {
		t.Errorf("Content mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedLines, newContent)
	}
}

func setupTestPythonPackage(t *testing.T, dir string) {
	// Create test_directives.py for testing
	testDirectives := `"""Test directives for testing."""
from typing import Optional
import time
import random

def process_ask(prompt: str) -> str:
    """Mock process_ask that returns a fixed response with random delay."""
    # Add random delay between 100ms and 300ms
    time.sleep(random.uniform(0.1, 0.3))
    return "Test response"

def process_do(action: str) -> str:
    """Mock process_do that returns a fixed response with random delay."""
    # Add random delay between 100ms and 300ms
    time.sleep(random.uniform(0.1, 0.3))
    return "Test response"
`
	if err := os.WriteFile(filepath.Join(dir, "test_directives.py"), []byte(testDirectives), 0644); err != nil {
		t.Fatal(err)
	}

	// Create src/pml/directives.py
	pmlDir := filepath.Join(dir, "src", "pml")
	if err := os.MkdirAll(pmlDir, 0755); err != nil {
		t.Fatal(err)
	}

	directivesContent := `import time
import random

def process_ask(prompt):
    # Add random delay between 100ms and 300ms
    time.sleep(random.uniform(0.1, 0.3))
    return "Test response"

def process_do(action):
    # Add random delay between 100ms and 300ms
    time.sleep(random.uniform(0.1, 0.3))
    return "Test response"
`
	if err := os.WriteFile(filepath.Join(pmlDir, "directives.py"), []byte(directivesContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create src/pml/__init__.py
	if err := os.WriteFile(filepath.Join(pmlDir, "__init__.py"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Create src/__init__.py
	if err := os.WriteFile(filepath.Join(dir, "src", "__init__.py"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Set PYTHONPATH to include the test directory
	if err := os.Setenv("PYTHONPATH", dir); err != nil {
		t.Fatal(err)
	}
}

func TestProcessFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "pml-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup Python package
	setupTestPythonPackage(t, tmpDir)

	// Create test PML file
	content := `:ask
What is 2+2?
:--

:do
Run some action
:--
`
	testFile := filepath.Join(tmpDir, "test.pml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create local .pml directory
	localResultsDir := filepath.Join(tmpDir, ".pml")
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create parser with mock LLM
	mockLLM := &mockLLM{
		response: "Test response",
		callback: func() {},
	}
	parser := NewParser(mockLLM, tmpDir, tmpDir, tmpDir)

	// Process the file
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	// Read the updated content
	updatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check if result links were added
	resultLinkPattern := regexp.MustCompile(`:-+\(r/[a-z]+_[a-z]+:"[^"]+"\)`)
	matches := resultLinkPattern.FindAllString(string(updatedContent), -1)
	if len(matches) != 2 {
		t.Errorf("Expected 2 result links, got %d\nContent:\n%s", len(matches), updatedContent)
	}

	// Verify each result link points to a valid file
	for _, match := range matches {
		resultName := strings.TrimPrefix(match, ":--(r/")
		resultName = strings.Split(resultName, ":")[0]
		resultPath := filepath.Join(localResultsDir, resultName+".pml")

		if _, err := os.Stat(resultPath); os.IsNotExist(err) {
			t.Errorf("Result file not found: %s", resultPath)
			continue
		}

		resultContent, err := os.ReadFile(resultPath)
		if err != nil {
			t.Errorf("Failed to read result file: %v", err)
			continue
		}

		if !strings.Contains(string(resultContent), "Test response") {
			t.Errorf("Result file missing expected content: %s", resultPath)
		}
	}

	// Verify Python file was created
	pyFile := testFile + ".py"
	if _, err := os.Stat(pyFile); os.IsNotExist(err) {
		t.Errorf("Expected Python file to be created: %s", pyFile)
	}
}

func TestUpdateContentWithResults(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "pml-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := `:ask
What is 2+2?
:--

:do
Run some action
:--
`
	blocks := []Block{
		{Type: ":ask", Content: []string{"What is 2+2?"}},
		{Type: ":do", Content: []string{"Run some action"}},
	}
	results := []string{"4", "Action completed"}

	// Create local .pml directory
	localResultsDir := filepath.Join(tmpDir, ".pml")
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create parser with mock LLM
	mockLLM := &mockLLM{
		response: "Test response",
		callback: func() {},
	}
	parser := NewParser(mockLLM, tmpDir, tmpDir, tmpDir)

	// Write result files first
	for i, block := range blocks {
		resultName := parser.generateUniqueResultName("test.pml", i, localResultsDir)
		resultFile := fmt.Sprintf("%s.pml", resultName)
		summary := fmt.Sprintf("Result for block %d from test.pml", i)
		if err := parser.writeResult(block, results[i], resultFile, localResultsDir, summary); err != nil {
			t.Fatalf("Failed to write result file: %v", err)
		}
	}

	// Update content with results
	updatedContent := parser.updateContentWithResults(blocks, content, results, localResultsDir, "test.pml")

	// Check if result links were added
	resultLinkPattern := regexp.MustCompile(`:-+\(r/[a-z]+_[a-z]+:"[^"]+"\)`)
	matches := resultLinkPattern.FindAllString(updatedContent, -1)
	if len(matches) != 2 {
		t.Errorf("Expected 2 result links, got %d\nContent:\n%s", len(matches), updatedContent)
	}

	// Verify each result link points to a valid file
	for i, match := range matches {
		resultName := strings.TrimPrefix(match, ":--(r/")
		resultName = strings.Split(resultName, ":")[0]
		resultPath := filepath.Join(localResultsDir, resultName+".pml")

		if _, err := os.Stat(resultPath); os.IsNotExist(err) {
			t.Errorf("Result file not found: %s", resultPath)
			continue
		}

		resultContent, err := os.ReadFile(resultPath)
		if err != nil {
			t.Errorf("Failed to read result file: %v", err)
			continue
		}

		expectedResult := results[i]
		if !strings.Contains(string(resultContent), expectedResult) {
			t.Errorf("Result file missing expected content: %s, got: %s", expectedResult, string(resultContent))
		}
	}
}

func TestCaching(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "pml-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup Python package
	setupTestPythonPackage(t, tmpDir)

	// Create test PML file
	content := `:ask
What is 2+2?
:--
`
	testFile := filepath.Join(tmpDir, "test.pml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create local .pml directory
	localResultsDir := filepath.Join(tmpDir, ".pml")
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Track number of times Ask is called
	var processCount int
	mockLLM := &mockLLM{
		response: "Test response",
		callback: func() {
			processCount++
		},
	}
	parser := NewParser(mockLLM, tmpDir, tmpDir, tmpDir)

	// First run - should process the block
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected 1 processing, got %d", processCount)
	}

	// Second run with same content - should use cache
	processCount = 0
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 0 {
		t.Errorf("Expected still 1 processing after rerun with same content, got %d", processCount)
	}

	// Change content - should process new block
	content = `:ask
What is 3+3?
:--
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	processCount = 0
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected 1 processing after content change, got %d", processCount)
	}

	// Clear cache - should process block again
	if err := os.RemoveAll(localResultsDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a new parser to clear in-memory cache
	parser = NewParser(mockLLM, tmpDir, tmpDir, tmpDir)

	processCount = 0
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected 1 processing after cache clear, got %d", processCount)
	}

	// Run again - should use cache
	processCount = 0
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 0 {
		t.Errorf("Expected still 1 processing (cache hit), got %d", processCount)
	}
}

func TestCacheCorruption(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "pml-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup Python package
	setupTestPythonPackage(t, tmpDir)

	// Create test PML file
	content := `:ask
What is 2+2?
:--
`
	testFile := filepath.Join(tmpDir, "test.pml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create local .pml directory
	localResultsDir := filepath.Join(tmpDir, ".pml")
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Track number of times Ask is called
	var processCount int
	mockLLM := &mockLLM{
		response: "Test response",
		callback: func() {
			processCount++
		},
	}
	parser := NewParser(mockLLM, tmpDir, tmpDir, tmpDir)

	// First run - should process the block
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected 1 processing with corrupted cache, got %d", processCount)
	}

	// Corrupt the cache file
	cacheFile := filepath.Join(localResultsDir, "cache.json")
	if err := os.WriteFile(cacheFile, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a new parser to clear in-memory cache
	parser = NewParser(mockLLM, tmpDir, tmpDir, tmpDir)

	// Run again - should process block again due to corrupted cache
	processCount = 0
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected 1 processing with corrupted cache, got %d", processCount)
	}

	// Run one more time - should use cache
	processCount = 0
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 0 {
		t.Errorf("Expected still 1 processing (cache hit), got %d", processCount)
	}
}

func TestParallelBlockProcessing(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "pml_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test environment
	setupTestPythonPackage(t, tempDir)
	os.Setenv("PYTHONPATH", tempDir)

	// Create test file with multiple blocks
	testFile := filepath.Join(tempDir, "test.pml")
	content := `:ask
What is 2+2?
:--

:ask
What is 3+3?
:--

:ask
What is 4+4?
:--

:ask
What is 5+5?
:--
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Process the file
	parser := NewParser(&mockLLM{response: "Test response"}, tempDir, filepath.Join(tempDir, "compiled"), filepath.Join(tempDir, "results"))
	startTime := time.Now()
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}
	totalDuration := time.Since(startTime)

	// Check if total duration is less than sum of individual delays
	// Each block takes 100-300ms, so 4 blocks in sequence would take at least 400ms
	// If parallel, it should take significantly less
	if totalDuration > 400*time.Millisecond {
		t.Errorf("Processing took too long (%v), suggesting sequential execution", totalDuration)
	}

	// Check if result files were created
	pmlDir := filepath.Join(filepath.Dir(testFile), ".pml")
	files, err := os.ReadDir(pmlDir)
	if err != nil {
		t.Fatal(err)
	}

	resultCount := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pml") {
			resultCount++
			// Read result file and verify content
			content, err := os.ReadFile(filepath.Join(pmlDir, file.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(content), "Test response") {
				t.Errorf("Result file %s does not contain expected response", file.Name())
			}
		}
	}

	if resultCount != 4 {
		t.Errorf("Expected 4 result files, got %d", resultCount)
	}
}

func TestProcessAllFiles(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "pml_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test environment
	setupTestPythonPackage(t, tempDir)
	os.Setenv("PYTHONPATH", tempDir)

	// Create test files
	files := []struct {
		name    string
		content string
	}{
		{
			name: "file1.pml",
			content: `:ask
What is 2+2?
:--

:ask
What is 3+3?
:--`,
		},
		{
			name: "file2.pml",
			content: `:ask
What is 4+4?
:--

:ask
What is 5+5?
:--`,
		},
		{
			name: "file3.pml",
			content: `:ask
What is 6+6?
:--

:ask
What is 7+7?
:--`,
		},
	}

	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tempDir, f.name), []byte(f.content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Process all files
	parser := NewParser(&mockLLM{response: "Test response"}, tempDir, filepath.Join(tempDir, "compiled"), filepath.Join(tempDir, "results"))
	startTime := time.Now()
	if err := parser.ProcessAllFiles(context.Background()); err != nil {
		t.Fatal(err)
	}
	totalDuration := time.Since(startTime)

	// Check if total duration is less than sum of individual delays
	// Each block takes 100-300ms, so 6 blocks in sequence would take at least 600ms
	// If parallel, it should take significantly less
	if totalDuration > 600*time.Millisecond {
		t.Errorf("Processing took too long (%v), suggesting sequential execution", totalDuration)
	}

	// Create .pml directory if it doesn't exist
	pmlDir := filepath.Join(tempDir, ".pml")
	if err := os.MkdirAll(pmlDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Check result files for each input file
	for _, f := range files {
		files, err := os.ReadDir(pmlDir)
		if err != nil {
			t.Fatal(err)
		}

		resultCount := 0
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".pml") {
				resultCount++
				// Read result file and verify content
				content, err := os.ReadFile(filepath.Join(pmlDir, file.Name()))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(content), "Test response") {
					t.Errorf("Result file %s does not contain expected response", file.Name())
				}
			}
		}

		if resultCount < 2 {
			t.Errorf("Expected at least 2 result files for %s, got %d", f.name, resultCount)
		}
	}
}
