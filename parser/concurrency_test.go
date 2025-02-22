package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestProcessAllFilesWithMixedContent tests concurrency with multiple PML files at once.
func TestProcessAllFilesWithMixedContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-conc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some PML files
	fileContents := []string{
		":ask\nHello?\n:--",
		":do\nRun that\n:--",
		":ask\nAnother question\n:--",
	}
	for i, c := range fileContents {
		err := os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("testfile_%c.pml", 'A'+i)), []byte(c), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Track LLM calls to verify concurrency
	var callCount int
	var mu sync.Mutex
	mockLLM := &mockLLM{
		response: "Test response",
		Delay:    50 * time.Millisecond, // 50ms delay so 3 in parallel take ~100ms total
		callback: func() {
			mu.Lock()
			callCount++
			mu.Unlock()
			// Optionally track calls here.
		},
	}

	// Create parser
	parser := NewParser(mockLLM, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	parser.SetForceProcess(true)

	start := time.Now()
	if err := parser.ProcessAllFiles(context.Background()); err != nil {
		t.Errorf("ProcessAllFiles concurrency test failed: %v", err)
	}
	dur := time.Since(start)

	// We expect 3 LLM calls (one for each file)
	if callCount != 3 {
		t.Errorf("Expected 3 LLM calls, got %d", callCount)
	}

	// If processing was sequential, it would take at least 300ms
	// With concurrency, it should be much less
	if dur > 200*time.Millisecond {
		t.Errorf("Expected concurrent processing, but duration suggests sequential: %v", dur)
	}
}

// TestConcurrentFileAccess tests that concurrent file access is handled properly.
func TestConcurrentFileAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-conc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file that multiple goroutines will try to process
	testFile := filepath.Join(tmpDir, "concurrent.pml")
	content := `:ask
What is 2+2?
:--`
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Track LLM calls
	var callCount int
	var mu sync.Mutex
	mockLLM := &mockLLM{
		response: "Test response",
		callback: func() {
			mu.Lock()
			callCount++
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
		},
	}

	parser := NewParser(mockLLM, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))

	// Process the same file concurrently
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := parser.ProcessFile(context.Background(), testFile)
			if err != nil {
				t.Errorf("Concurrent ProcessFile failed: %v", err)
			}
		}()
	}

	wg.Wait()

	// Due to caching, we should only see one LLM call
	if callCount != 1 {
		t.Errorf("Expected 1 LLM call due to caching, got %d", callCount)
	}
}

// TestProcessAllFilesCancellation tests that processing can be cancelled.
func TestProcessAllFilesCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-conc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create several files that will take time to process
	for i := 0; i < 10; i++ {
		content := fmt.Sprintf(`:ask
Question %c
:--`, 'A'+i)
		err := os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%c.pml", 'A'+i)), []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create a mock LLM that takes time to respond
	var processedCount int
	var mu sync.Mutex
	mockLLM := &mockLLM{
		response: "Test response",
		Delay:    500 * time.Millisecond, // 500ms delay so processing exceeds the 250ms deadline
		callback: func() {
			mu.Lock()
			processedCount++
			mu.Unlock()
			// Optionally track calls here.
		},
	}

	parser := NewParser(mockLLM, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))

	// Create a context that will be cancelled shortly
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Start processing
	err = parser.ProcessAllFiles(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected deadline exceeded error, got: %v", err)
	}

	// We should have processed some but not all files
	if processedCount == 0 {
		t.Error("Expected some files to be processed before cancellation")
	}
	if processedCount == 10 {
		t.Error("Expected processing to be cancelled before completion")
	}
}
