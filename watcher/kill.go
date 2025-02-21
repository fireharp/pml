package watcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ResultsWatcher watches for file system changes in the results directory and kills processes writing to it
type ResultsWatcher struct {
	watchPath string
	fsWatcher *fsnotify.Watcher
	done      chan struct{}
}

// NewResultsWatcher creates a new watcher for the results directory
func NewResultsWatcher(resultsDir string) (*ResultsWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	if err := fsWatcher.Add(resultsDir); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("failed to add watch path: %w", err)
	}

	w := &ResultsWatcher{
		watchPath: resultsDir,
		fsWatcher: fsWatcher,
		done:      make(chan struct{}),
	}

	// Write PID file
	if err := w.writePidFile(); err != nil {
		log.Printf("Warning: Failed to write PID file: %v", err)
	}

	return w, nil
}

// getPidFileName creates a unique PID filename for the results watcher
func (w *ResultsWatcher) getPidFileName() string {
	safePath := strings.ReplaceAll(w.watchPath, string(filepath.Separator), "_")
	return fmt.Sprintf("pml-results-killer-%s.pid", safePath)
}

// writePidFile writes the current process PID to a file
func (w *ResultsWatcher) writePidFile() error {
	pidDir, err := getPidDir()
	if err != nil {
		return err
	}
	pidFile := filepath.Join(pidDir, w.getPidFileName())
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

// removePidFile removes the PID file for this watcher
func (w *ResultsWatcher) removePidFile() {
	if pidDir, err := getPidDir(); err == nil {
		pidFile := filepath.Join(pidDir, w.getPidFileName())
		_ = os.Remove(pidFile)
	}
}

// Start begins watching the results directory and killing processes that write to it
func (w *ResultsWatcher) Start() {
	log.Printf("Starting results watcher for %s\n", w.watchPath)

	// Verify the directory exists
	if _, err := os.Stat(w.watchPath); err != nil {
		log.Printf("Warning: Results directory does not exist: %v\n", err)
		if err := os.MkdirAll(w.watchPath, 0755); err != nil {
			log.Printf("Failed to create results directory: %v\n", err)
			return
		}
	}

	// Keep watching until explicitly stopped
	for {
		select {
		case <-w.done:
			log.Printf("Received done signal, stopping watcher\n")
			w.removePidFile() // Remove PID file when stopping
			return
		default:
			// Re-add the watch path in case it was removed
			if err := w.fsWatcher.Add(w.watchPath); err != nil {
				log.Printf("Error re-adding watch path: %v\n", err)
				time.Sleep(time.Second) // Wait before retrying
				continue
			}

			// Process events
			select {
			case <-w.done:
				log.Printf("Received done signal, stopping watcher\n")
				return
			case event, ok := <-w.fsWatcher.Events:
				if !ok {
					log.Printf("Event channel closed, restarting watcher\n")
					time.Sleep(time.Second) // Wait before retrying
					continue
				}
				// Check for write or create events
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					log.Printf("Detected modification in: %s (op: %v)\n", event.Name, event.Op)
					if _, err := os.Stat(event.Name); err != nil {
						log.Printf("Warning: File no longer exists: %v\n", err)
						continue
					}
					if err := w.killWritingProcesses(event.Name); err != nil {
						log.Printf("Error killing processes: %v\n", err)
					}
				}
			case err, ok := <-w.fsWatcher.Errors:
				if !ok {
					log.Printf("Error channel closed, restarting watcher\n")
					time.Sleep(time.Second) // Wait before retrying
					continue
				}
				log.Printf("Watcher error: %v\n", err)
			}
		}
	}
}

// Stop stops the watcher
func (w *ResultsWatcher) Stop() error {
	close(w.done)
	w.removePidFile()
	return w.fsWatcher.Close()
}

// killWritingProcesses finds and kills processes writing to the specified file
func (w *ResultsWatcher) killWritingProcesses(filePath string) error {
	log.Printf("Looking for processes writing to: %s\n", filePath)
	currentPid := os.Getpid()

	// Keep trying to kill processes until none are found
	for attempts := 0; attempts < 5; attempts++ {
		// Use lsof to find processes accessing the file, but only look for write access
		cmd := exec.Command("lsof", "-w", "-F", "pc", filePath) // -F pc gives us PID and command in machine format
		output, err := cmd.Output()
		if err != nil {
			// If lsof returns no results, that's not an error
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				log.Printf("No processes found writing to: %s\n", filePath)
				return nil
			}
			return fmt.Errorf("error executing lsof: %w", err)
		}

		// Parse the machine-formatted output to find process IDs
		var killedPids []string
		lines := strings.Split(string(output), "\n")
		log.Printf("lsof raw output:\n%s\n", output)

		foundProcesses := false
		var currentCmd string
		for _, line := range lines {
			if line == "" {
				continue
			}
			// Lines starting with 'p' contain the PID
			// Lines starting with 'c' contain the command name
			switch line[0] {
			case 'p':
				pid := line[1:] // Skip the 'p' prefix
				pidInt, err := strconv.Atoi(pid)
				if err != nil {
					log.Printf("Invalid PID %s: %v\n", pid, err)
					continue
				}

				// Skip our own process and any child processes (like lsof)
				if pidInt == currentPid {
					log.Printf("Skipping our own process: %d (%s)\n", pidInt, currentCmd)
					continue
				}

				// Check if this is a parent process of ours
				if isAncestorProcess(pidInt) {
					log.Printf("Skipping ancestor process: %d (%s)\n", pidInt, currentCmd)
					continue
				}

				foundProcesses = true
				log.Printf("Attempting to terminate process: %s (%s)\n", pid, currentCmd)
				if err := terminateProcess(pid); err != nil {
					log.Printf("Failed to terminate process %s: %v\n", pid, err)
				} else {
					killedPids = append(killedPids, fmt.Sprintf("%s(%s)", pid, currentCmd))
					log.Printf("Successfully terminated process: %s (%s)\n", pid, currentCmd)
				}
			case 'c':
				currentCmd = line[1:] // Skip the 'c' prefix
			}
		}

		if len(killedPids) > 0 {
			log.Printf("Killed processes writing to %s: %v\n", filePath, killedPids)
		}

		// If no processes were found, we can stop trying
		if !foundProcesses {
			log.Printf("No more processes found writing to: %s\n", filePath)
			return nil
		}

		// Wait a bit before checking again
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// isAncestorProcess checks if the given PID is an ancestor of our process
func isAncestorProcess(pid int) bool {
	currentPid := os.Getpid()
	for currentPid != 1 { // 1 is the init process
		ppid, err := getParentPID(currentPid)
		if err != nil {
			return false
		}
		if ppid == pid {
			return true
		}
		currentPid = ppid
	}
	return false
}

// getParentPID gets the parent PID of a process
func getParentPID(pid int) (int, error) {
	cmd := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	ppid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}
	return ppid, nil
}

// terminateProcess terminates a process by its PID
func terminateProcess(pid string) error {
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Errorf("invalid PID: %w", err)
	}

	// First try SIGTERM for graceful shutdown
	proc, err := os.FindProcess(pidInt)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	// Try SIGTERM first
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		log.Printf("SIGTERM failed for PID %d, trying SIGKILL: %v\n", pidInt, err)
		// If SIGTERM fails, use SIGKILL
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	// Give the process a moment to terminate gracefully
	time.Sleep(100 * time.Millisecond)
	return nil
}

// CleanupResultsWatchers kills all running results watchers and removes their PID files
func CleanupResultsWatchers() error {
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
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "pml-results-killer-") {
			pidFile := filepath.Join(pidDir, entry.Name())
			if pidBytes, err := os.ReadFile(pidFile); err == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes))); err == nil {
					// Skip current process
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
	return nil
}
