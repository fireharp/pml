package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockProcessor is a mock file processor for testing
type mockProcessor struct {
	mu       sync.Mutex
	files    []string
	err      error
	callback func(string) // optional callback for custom verification
}

func (m *mockProcessor) ProcessFile(_ context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files = append(m.files, path)
	if m.callback != nil {
		m.callback(path)
	}
	return m.err
}

func TestWatcher(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "watcher-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create processor with a channel to signal when processing is done
	processed := make(chan string, 1)
	processor := &mockProcessor{
		callback: func(path string) {
			processed <- path
		},
	}

	// Create watcher
	w, err := NewWatcher(tmpDir, processor)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Start watcher in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := w.Start(ctx); err != nil && err != context.Canceled {
			t.Errorf("Watcher.Start() error = %v", err)
		}
	}()

	// Wait for watcher to start
	time.Sleep(100 * time.Millisecond)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait a bit and then modify the file to trigger processing
	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for processing
	select {
	case processedFile := <-processed:
		if processedFile != testFile {
			t.Errorf("Processed file = %v, want %v", processedFile, testFile)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for file to be processed")
	}

	// Clean up
	cancel()
	wg.Wait()
}

func TestWatcherErrors(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		processor FileProcessor
		wantErr   bool
	}{
		{
			name:      "Invalid path",
			path:      "/nonexistent/path",
			processor: &mockProcessor{},
			wantErr:   true,
		},
		{
			name:      "Nil processor",
			path:      ".",
			processor: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewWatcher(tt.path, tt.processor)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				w.Close()
			}
		})
	}
}

func TestPidFileManagement(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "watcher-pid-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Clean up any existing PID files before starting
	if err := CleanupWatchers(); err != nil {
		t.Fatal(err)
	}

	// Create a mock processor
	processor := &mockProcessor{}

	// Create watcher
	w, err := NewWatcher(tmpDir, processor)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	// Start watcher in background
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := w.Start(ctx); err != nil && err != context.Canceled {
			t.Errorf("Watcher.Start() error = %v", err)
		}
	}()

	// Wait for watcher to start
	time.Sleep(100 * time.Millisecond)

	// Check if PID file exists and contains correct PID
	pidDir, err := getPidDir()
	if err != nil {
		cancel()
		wg.Wait()
		t.Fatal(err)
	}

	pidFile := filepath.Join(pidDir, getPidFileName(tmpDir))
	if _, err := os.Stat(pidFile); err != nil {
		cancel()
		wg.Wait()
		t.Errorf("PID file not created: %v", err)
	}

	// Read PID file and verify content
	pidBytes, err := os.ReadFile(pidFile)
	if err != nil {
		cancel()
		wg.Wait()
		t.Fatal(err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		cancel()
		wg.Wait()
		t.Fatal(err)
	}

	if pid != os.Getpid() {
		cancel()
		wg.Wait()
		t.Errorf("PID file contains wrong PID. got = %d, want = %d", pid, os.Getpid())
	}

	// Clean up
	cancel()
	wg.Wait()

	// Verify PID file is removed after watcher stops
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Error("PID file not removed after watcher stops")
	}
}

func TestMultipleWatchers(t *testing.T) {
	// Clean up any existing PID files before starting
	if err := CleanupWatchers(); err != nil {
		t.Fatal(err)
	}

	// Create temporary directories for multiple watchers
	tmpDirs := make([]string, 3)
	for i := range tmpDirs {
		dir, err := os.MkdirTemp("", fmt.Sprintf("watcher-multi-test-%d-*", i))
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)
		tmpDirs[i] = dir
	}

	// Create watchers
	watchers := make([]*Watcher, len(tmpDirs))
	for i, dir := range tmpDirs {
		w, err := NewWatcher(dir, &mockProcessor{})
		if err != nil {
			t.Fatal(err)
		}
		watchers[i] = w
		defer w.Close()
	}

	// Start all watchers
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for _, w := range watchers {
		wg.Add(1)
		go func(w *Watcher) {
			defer wg.Done()
			if err := w.Start(ctx); err != nil && err != context.Canceled {
				t.Errorf("Watcher.Start() error = %v", err)
			}
		}(w)
	}

	// Wait for watchers to start
	time.Sleep(100 * time.Millisecond)

	// Verify PID files exist for all watchers
	pidDir, err := getPidDir()
	if err != nil {
		cancel()
		wg.Wait()
		t.Fatal(err)
	}

	entries, err := os.ReadDir(pidDir)
	if err != nil {
		cancel()
		wg.Wait()
		t.Fatal(err)
	}

	pidFiles := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "pml-watcher-") {
			pidFiles++
		}
	}

	if pidFiles != len(tmpDirs) {
		cancel()
		wg.Wait()
		t.Errorf("Wrong number of PID files. got = %d, want = %d", pidFiles, len(tmpDirs))
	}

	// Clean up
	cancel()
	wg.Wait()

	// Verify all PID files are removed after watchers stop
	entries, err = os.ReadDir(pidDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	remainingPidFiles := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "pml-watcher-") {
			remainingPidFiles++
		}
	}

	if remainingPidFiles != 0 {
		t.Errorf("PID files remain after watchers stop. got = %d, want = 0", remainingPidFiles)
	}
}

func TestCleanupNonExistentProcesses(t *testing.T) {
	// Clean up any existing PID files before starting
	if err := CleanupWatchers(); err != nil {
		t.Fatal(err)
	}

	// Create a temporary PID directory
	pidDir, err := getPidDir()
	if err != nil {
		t.Fatal(err)
	}

	// Create fake PID files with non-existent PIDs
	fakePids := []int{999999, 999998, 999997} // Using high PIDs that are unlikely to exist
	for i, pid := range fakePids {
		pidFile := filepath.Join(pidDir, fmt.Sprintf("pml-watcher-fake-%d.pid", i))
		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Run cleanup
	if err := CleanupWatchers(); err != nil {
		t.Errorf("CleanupWatchers() error = %v", err)
	}

	// Verify all PID files are removed
	entries, err := os.ReadDir(pidDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	remainingPidFiles := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "pml-watcher-") {
			remainingPidFiles++
		}
	}

	if remainingPidFiles != 0 {
		t.Errorf("PID files remain after cleanup. got = %d, want = 0", remainingPidFiles)
	}
}
