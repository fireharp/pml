package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	pidDirName = ".pml/watchers"
)

// FileEvent represents a file system event
type FileEvent struct {
	Type      string    `json:"type"`
	File      string    `json:"file"`
	Timestamp time.Time `json:"timestamp"`
}

// FileProcessor interface for processing files
type FileProcessor interface {
	ProcessFile(ctx context.Context, path string) error
}

// Watcher watches for file system changes
type Watcher struct {
	watchPath string
	fsWatcher *fsnotify.Watcher
	processor FileProcessor
}

// NewWatcher creates a new file system watcher
func NewWatcher(watchPath string, processor FileProcessor) (*Watcher, error) {
	if processor == nil {
		return nil, fmt.Errorf("processor cannot be nil")
	}

	absPath, err := filepath.Abs(watchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("watch path does not exist: %w", err)
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Watcher{
		watchPath: absPath,
		fsWatcher: fsWatcher,
		processor: processor,
	}, nil
}

// getPidDir returns the directory where PID files are stored
func getPidDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	pidDir := filepath.Join(homeDir, pidDirName)
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create PID directory: %w", err)
	}
	return pidDir, nil
}

// getPidFileName creates a unique PID filename for a workspace
func getPidFileName(workspaceDir string) string {
	safePath := strings.ReplaceAll(workspaceDir, string(filepath.Separator), "_")
	return fmt.Sprintf("pml-watcher-%s.pid", safePath)
}

// writePidFile writes the current process PID to a file
func (w *Watcher) writePidFile() error {
	pidDir, err := getPidDir()
	if err != nil {
		return err
	}
	pidFile := filepath.Join(pidDir, getPidFileName(w.watchPath))
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

// removePidFile removes the PID file for this watcher
func (w *Watcher) removePidFile() {
	if pidDir, err := getPidDir(); err == nil {
		pidFile := filepath.Join(pidDir, getPidFileName(w.watchPath))
		_ = os.Remove(pidFile)
	}
}

// CleanupWatchers kills all running watchers and removes their PID files
func CleanupWatchers() error {
	pidDir, err := getPidDir()
	if err != nil {
		return fmt.Errorf("failed to get PID directory: %w", err)
	}

	entries, err := os.ReadDir(pidDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No watchers to clean up
		}
		return fmt.Errorf("failed to read PID directory: %w", err)
	}

	currentPid := os.Getpid()
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "pml-watcher-") {
			pidFile := filepath.Join(pidDir, entry.Name())
			if pidBytes, err := os.ReadFile(pidFile); err == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes))); err == nil {
					// Skip current process during tests
					if pid == currentPid {
						continue
					}
					if proc, err := os.FindProcess(pid); err == nil {
						err := proc.Kill()
						if err != nil && !strings.Contains(err.Error(), "process already finished") {
							log.Printf("Warning: failed to kill process %d: %v", pid, err)
						}
					}
				}
			}
			// Always remove PID file, regardless of whether we could kill the process
			if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
				log.Printf("Warning: failed to remove PID file %s: %v", pidFile, err)
			}
		}
	}
	return nil // Return nil since "process already finished" is not a real error
}

// Start starts watching for file system events
func (w *Watcher) Start(ctx context.Context) error {
	// Write PID file when starting
	if err := w.writePidFile(); err != nil {
		log.Printf("Warning: Failed to write PID file: %v", err)
	}
	defer w.removePidFile()

	// Add the path to watch
	if err := w.fsWatcher.Add(w.watchPath); err != nil {
		return fmt.Errorf("failed to add watch path: %w", err)
	}

	fmt.Printf("PML-INIT: Starting watcher for %s\n", w.watchPath)

	// Start listening for events
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return fmt.Errorf("watcher event channel closed")
			}

			// Debounce write events
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Skip if the file is being written to avoid processing partial writes
				continue
			}

			// Create structured event
			fileEvent := FileEvent{
				Type:      w.getEventType(event.Op),
				File:      event.Name,
				Timestamp: time.Now(),
			}

			// Output JSON for VSCode to consume
			if jsonData, err := json.Marshal(fileEvent); err == nil {
				fmt.Printf("PML-EVENT: %s\n", string(jsonData))
			}

			// Process file if it was created or closed after writing
			if event.Op&(fsnotify.Create|fsnotify.Chmod) != 0 {
				if err := w.processor.ProcessFile(ctx, event.Name); err != nil {
					log.Printf("Failed to process file: %v", err)
				}
			}

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return fmt.Errorf("watcher error channel closed")
			}
			log.Printf("Watcher error: %v", err)
			errorEvent := FileEvent{
				Type:      "error",
				File:      "",
				Timestamp: time.Now(),
			}
			if jsonData, err := json.Marshal(errorEvent); err == nil {
				fmt.Printf("PML-ERROR: %s\n", string(jsonData))
			}
		}
	}
}

// Close closes the watcher
func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}

// getEventType converts fsnotify operation to string
func (w *Watcher) getEventType(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "created"
	case op&fsnotify.Write == fsnotify.Write:
		return "modified"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "deleted"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "renamed"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "chmod"
	default:
		return "unknown"
	}
}
