package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ProcessAllFiles processes all PML files in the source directory concurrently
func (p *Parser) ProcessAllFiles(ctx context.Context) error {
	if err := p.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	files, err := p.findPMLFiles()
	if err != nil {
		return fmt.Errorf("failed to find PML files: %w", err)
	}

	// Create a channel for results
	results := make(chan error, len(files))
	var wg sync.WaitGroup

	// Process each file concurrently
	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			if err := p.ProcessFile(ctx, f); err != nil {
				results <- fmt.Errorf("error processing %s: %w", f, err)
				return
			}
			results <- nil
		}(file)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect errors
	var errors []error
	for err := range results {
		if err != nil {
			errors = append(errors, err)
		}
	}
	if ctx.Err() != nil {
	    return ctx.Err()
	}

	if len(errors) > 0 {
		// Format all errors into a single error
		errStr := fmt.Sprintf("encountered %d errors:\n", len(errors))
		for _, err := range errors {
			errStr += fmt.Sprintf("- %v\n", err)
		}
		return fmt.Errorf(errStr)
	}

	return nil
}

// findPMLFiles finds all PML files in the source directory
func (p *Parser) findPMLFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(p.sourcesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && IsPMLFile(path) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// ensureDirectories creates necessary directories if they don't exist
func (p *Parser) ensureDirectories() error {
	dirs := []string{p.sourcesDir, p.compiledDir, p.rootResultsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}
