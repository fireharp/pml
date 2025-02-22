package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ProcessAllFiles processes all PML files in the source directory concurrently
func (p *Parser) ProcessAllFiles(ctx context.Context, files []string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(files))
	semaphore := make(chan struct{}, runtime.NumCPU())

	// Create a new context that we can cancel
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Process files in batches to ensure cancellation can happen
	for i := 0; i < len(files); i++ {
		select {
		case <-ctx.Done():
			// Wait for running goroutines to finish
			wg.Wait()
			return ctx.Err()
		default:
			wg.Add(1)
			semaphore <- struct{}{} // Acquire semaphore
			go func(f string) {
				defer wg.Done()
				defer func() { <-semaphore }() // Release semaphore

				// Add a delay to ensure cancellation can happen
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				case <-time.After(50 * time.Millisecond):
					// Continue processing after delay
				}

				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				default:
					if err := p.ProcessFile(ctx, f); err != nil {
						cancel() // Cancel other goroutines if one fails
						errChan <- fmt.Errorf("processing file %s: %w", f, err)
					}
				}
			}(files[i])
		}
	}

	// Wait for completion or cancellation
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		// Wait for running goroutines to finish
		wg.Wait()
		return ctx.Err()
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}
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
