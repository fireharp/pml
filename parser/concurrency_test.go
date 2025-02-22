package parser

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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
	var files []string
	for i, c := range fileContents {
		f := filepath.Join(tmpDir, fmt.Sprintf("testfile_%c.pml", 'A'+i))
		err := os.WriteFile(f, []byte(c), 0644)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
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
	if err := parser.ProcessAllFiles(context.Background(), files); err != nil {
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
	tmpDir, err := os.MkdirTemp("", "pml-concurrency-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files with blocks that will trigger LLM calls
	files := []string{
		filepath.Join(tmpDir, "test1.pml"),
		filepath.Join(tmpDir, "test2.pml"),
		filepath.Join(tmpDir, "test3.pml"),
	}
	for i, f := range files {
		content := fmt.Sprintf(":ask\nQuestion %d\n:--", i+1)
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var callCount int32
	parser := NewParser(&mockLLM{
		response: "Test response",
		Delay:    100 * time.Millisecond,
		callback: func() {
			atomic.AddInt32(&callCount, 1)
		},
	}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))

	parser.SetForceProcess(true)

	start := time.Now()
	if err := parser.ProcessAllFiles(context.Background(), files); err != nil {
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

// TestProcessAllFilesCancellation tests that processing can be cancelled.
func TestProcessAllFilesCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-cancel-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files with blocks that will trigger LLM calls
	var files []string
	for i := 0; i < 10; i++ {
		f := filepath.Join(tmpDir, fmt.Sprintf("test%d.pml", i))
		content := fmt.Sprintf(":ask\nQuestion %d\n:--", i+1)
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}

	var processedCount int32
	parser := NewParser(&mockLLM{
		response: "Test response",
		Delay:    100 * time.Millisecond,
		callback: func() {
			atomic.AddInt32(&processedCount, 1)
			time.Sleep(50 * time.Millisecond) // Add extra delay to ensure cancellation
		},
	}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = parser.ProcessAllFiles(ctx, files)
	if err == nil {
		t.Error("Expected error due to cancellation")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected deadline exceeded error, got: %v", err)
	}

	// We should have processed some files
	count := atomic.LoadInt32(&processedCount)
	if count == 0 {
		t.Error("Expected some files to be processed before cancellation")
	}
}
