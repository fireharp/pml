package parser

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
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

func setupTestPythonPackage(t *testing.T, baseDir string) {
	// Create src/pml/directives package structure
	pkgPath := filepath.Join(baseDir, "src", "pml", "directives")
	if err := os.MkdirAll(pkgPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create __init__.py files
	initFiles := []string{
		filepath.Join(baseDir, "src", "__init__.py"),
		filepath.Join(baseDir, "src", "pml", "__init__.py"),
	}
	for _, file := range initFiles {
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create directives/__init__.py with test functions
	directivesInit := `"""Test directives for testing."""

def process_ask(prompt: str) -> str:
    """Mock process_ask that returns a fixed response."""
    return "Test response"

def process_do(action: str) -> str:
    """Mock process_do that returns a fixed response."""
    return "Test response"
`
	if err := os.WriteFile(filepath.Join(pkgPath, "__init__.py"), []byte(directivesInit), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test_directives.py for testing
	testDirectives := `"""Test directives for testing."""
from typing import Optional

def process_ask(prompt: str) -> str:
    """Mock process_ask that returns a fixed response."""
    return "Test response"

def process_do(action: str) -> str:
    """Mock process_do that returns a fixed response."""
    return "Test response"
`
	if err := os.WriteFile(filepath.Join(baseDir, "test_directives.py"), []byte(testDirectives), 0644); err != nil {
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

	// Create directory structure - using same directory for sources and compiled
	sourcesDir := filepath.Join(tmpDir, "sources")
	resultsDir := filepath.Join(tmpDir, "results")

	// Create directories
	for _, dir := range []string{sourcesDir, resultsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Setup Python package in the sources directory
	setupTestPythonPackage(t, sourcesDir)

	// Create test PML file
	testFile := filepath.Join(sourcesDir, "test.pml")
	if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
		t.Fatal(err)
	}

	// Create local .pml directory
	localResultsDir := filepath.Join(filepath.Dir(testFile), ".pml")
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		t.Fatal(err)
	}

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
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create parser with mock LLM - using sourcesDir for both source and compiled
	mockLLM := &mockLLM{response: "Test response"}
	parser := NewParser(mockLLM, sourcesDir, sourcesDir, resultsDir)

	// Process file
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	// Verify Python file was created in the same directory
	pyFile := testFile + ".py"
	if _, err := os.Stat(pyFile); os.IsNotExist(err) {
		t.Error("Expected Python file to be created in the same directory")
	}

	pyContent, err := os.ReadFile(pyFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check Python code content - should be exact copy with blocks replaced
	expectedParts := []string{
		"# Auto-generated imports for PML blocks",
		"from src.pml.directives import process_ask, process_do",
		"# :ask",
		"result_0 = process_ask('''",
		"What is 2+2?",
		"''')",
		"# :--",
		"def some_code():",
		"    print(\"hello\")",
		"# :do",
		"result_1 = process_do('''",
		"Run some action",
		"''')",
		"# :--",
		"some_code()",
	}

	for _, part := range expectedParts {
		if !strings.Contains(string(pyContent), part) {
			t.Errorf("Python file missing expected part: %s", part)
		}
	}

	// Verify block execution files were created and cleaned up
	blockFiles := []string{
		filepath.Join(sourcesDir, ".test.pml.block_0.py"),
		filepath.Join(sourcesDir, ".test.pml.block_1.py"),
	}
	for _, blockFile := range blockFiles {
		if _, err := os.Stat(blockFile); !os.IsNotExist(err) {
			t.Errorf("Block file should have been cleaned up: %s", blockFile)
		}
	}

	// Verify PML file was updated with result links
	updatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Check if result links were added and are unique
	resultLinkPattern := regexp.MustCompile(`:-+\(r/[a-z]+_[a-z]+\)`)
	matches := resultLinkPattern.FindAllString(string(updatedContent), -1)
	if len(matches) != 2 {
		t.Errorf("Expected 2 result links, got %d\nContent:\n%s", len(matches), string(updatedContent))
	}

	// Verify each result link is unique
	seenLinks := make(map[string]bool)
	for _, match := range matches {
		if seenLinks[match] {
			t.Errorf("Found duplicate result link: %s", match)
		}
		seenLinks[match] = true
	}

	// Verify result files exist and contain correct content
	for _, match := range matches {
		resultName := strings.TrimPrefix(match, ":--(r/")
		resultName = strings.TrimSuffix(resultName, ")")
		resultPath := filepath.Join(filepath.Dir(testFile), ".pml", resultName+".pml")

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

	parser := NewParser(nil, "", "", tmpDir)
	updatedContent := parser.updateContentWithResults(blocks, content, results, localResultsDir, "test.pml")

	// Check if result links were added and are unique
	resultLinkPattern := regexp.MustCompile(`:-+\(r/[a-z]+_[a-z]+\)`)
	matches := resultLinkPattern.FindAllString(updatedContent, -1)
	if len(matches) != 2 {
		t.Errorf("Expected 2 result links, got %d\nContent:\n%s", len(matches), updatedContent)
	}

	// Verify each result link is unique
	seenLinks := make(map[string]bool)
	for _, match := range matches {
		if seenLinks[match] {
			t.Errorf("Found duplicate result link: %s", match)
		}
		seenLinks[match] = true
	}

	// Verify result files exist and are unique
	seenFiles := make(map[string]bool)
	for _, match := range matches {
		resultName := strings.TrimPrefix(match, ":--(r/")
		resultName = strings.TrimSuffix(resultName, ")")
		resultPath := filepath.Join(localResultsDir, resultName+".pml")

		if _, err := os.Stat(resultPath); os.IsNotExist(err) {
			t.Errorf("Result file not found: %s", resultPath)
		}

		if seenFiles[resultPath] {
			t.Errorf("Found duplicate result file: %s", resultPath)
		}
		seenFiles[resultPath] = true
	}
}

func TestCaching(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "pml-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure - using same directory for sources and compiled
	sourcesDir := filepath.Join(tmpDir, "sources")
	resultsDir := filepath.Join(tmpDir, "results")

	// Create directories
	for _, dir := range []string{sourcesDir, resultsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Setup Python package in the sources directory
	setupTestPythonPackage(t, sourcesDir)

	// Create test PML file
	testFile := filepath.Join(sourcesDir, "test.pml")
	if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
		t.Fatal(err)
	}

	// Create local .pml directory
	localResultsDir := filepath.Join(filepath.Dir(testFile), ".pml")
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Track number of times Ask is called
	processCount := 0
	mockLLM := &mockLLM{
		response: "Test response",
		callback: func() {
			processCount++
		},
	}

	// Create parser with mock LLM - using sourcesDir for both source and compiled
	parser := NewParser(mockLLM, sourcesDir, sourcesDir, resultsDir)

	// Test Case 1: Initial processing
	content := `:ask
What is 2+2?
:--
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected 1 processing, got %d", processCount)
	}

	// Verify cache file was created in .pml directory
	cacheFile := filepath.Join(sourcesDir, ".pml", "cache.json")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Error("Expected cache file to be created in .pml directory")
	}

	// Test Case 2: Same content with different result link
	// Read the current content to verify it has a result link
	updatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Verify result link was added
	if !strings.Contains(string(updatedContent), ":--(r/") {
		t.Error("Expected result link to be added to the file")
	}

	// Process again - should not increase process count despite having result link
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected still 1 processing after rerun with same content, got %d", processCount)
	}

	// Test Case 3: Modified content
	newContent := `:ask
What is 3+3?
:--
`
	if err := os.WriteFile(testFile, []byte(newContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 2 {
		t.Errorf("Expected 2 processings after content change, got %d", processCount)
	}

	// Test Case 4: Clear cache and reprocess same content
	if err := os.Remove(cacheFile); err != nil {
		t.Fatal(err)
	}

	// Reload parser to clear in-memory cache
	parser = NewParser(mockLLM, sourcesDir, sourcesDir, resultsDir)

	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 3 {
		t.Errorf("Expected 3 processings after cache clear, got %d", processCount)
	}

	// Test Case 5: Reprocess immediately after - should use cache
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 3 {
		t.Errorf("Expected still 3 processings (cache hit), got %d", processCount)
	}

	// Verify Python files are created in the same directory
	pyFile := testFile + ".py"
	if _, err := os.Stat(pyFile); os.IsNotExist(err) {
		t.Error("Expected Python file to be created in the same directory")
	}
}

func TestCacheCorruption(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "pml-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure - using same directory for sources and compiled
	sourcesDir := filepath.Join(tmpDir, "sources")
	resultsDir := filepath.Join(tmpDir, "results")

	// Create directories
	for _, dir := range []string{sourcesDir, resultsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Setup Python package in the sources directory
	setupTestPythonPackage(t, sourcesDir)

	// Create .pml directory and corrupted cache file
	pmlDir := filepath.Join(sourcesDir, ".pml")
	if err := os.MkdirAll(pmlDir, 0755); err != nil {
		t.Fatal(err)
	}
	cacheFile := filepath.Join(pmlDir, "cache.json")
	if err := os.WriteFile(cacheFile, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test file
	testFile := filepath.Join(sourcesDir, "test.pml")
	if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
		t.Fatal(err)
	}

	// Create local .pml directory
	localResultsDir := filepath.Join(filepath.Dir(testFile), ".pml")
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `:ask
What is 2+2?
:--
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Track number of times Ask is called
	processCount := 0
	mockLLM := &mockLLM{
		response: "Test response",
		callback: func() {
			processCount++
		},
	}

	// Create parser with mock LLM - using sourcesDir for both source and compiled
	parser := NewParser(mockLLM, sourcesDir, sourcesDir, resultsDir)

	// Should still process file despite corrupted cache
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected 1 processing with corrupted cache, got %d", processCount)
	}

	// Verify cache was recreated correctly
	cacheData, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatal(err)
	}

	var cache map[string]CacheEntry
	if err := json.Unmarshal(cacheData, &cache); err != nil {
		t.Error("Expected cache file to be valid JSON after processing")
	}

	// Verify cache entry exists and is correct
	if entry, ok := cache[testFile]; !ok {
		t.Error("Expected cache entry for test file")
	} else {
		expectedChecksum := parser.calculateChecksum(content)
		if entry.Checksum != expectedChecksum {
			t.Errorf("Cache checksum = %v, want %v", entry.Checksum, expectedChecksum)
		}
	}

	// Process again - should use cache
	if err := parser.ProcessFile(context.Background(), testFile); err != nil {
		t.Fatal(err)
	}

	if processCount != 1 {
		t.Errorf("Expected still 1 processing (cache hit), got %d", processCount)
	}
}
