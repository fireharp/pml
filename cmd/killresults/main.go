package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fireharp/pml/impl1/watcher"
)

func main() {
	// Set up logging
	log.SetFlags(0)

	// Parse command line flags
	workspaceDir := flag.String("dir", ".", "Workspace directory containing the results folder")
	flag.Parse()

	// Clean up any existing watchers
	if err := watcher.CleanupResultsWatchers(); err != nil {
		log.Printf("Warning: Failed to clean up existing watchers: %v", err)
	}

	// Get absolute path of workspace directory
	absWorkspaceDir, err := filepath.Abs(*workspaceDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Set up results directory path
	resultsDir := filepath.Join(absWorkspaceDir, "impl1", "results")

	// Create results watcher
	w, err := watcher.NewResultsWatcher(resultsDir)
	if err != nil {
		log.Fatalf("Failed to create results watcher: %v", err)
	}
	defer w.Stop()

	// Start watching
	w.Start()
	log.Printf("Started watching %s for file modifications\n", resultsDir)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
