package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ProcessFile processes a single PML file (parse, generate .py, run blocks in parallel)
func (p *Parser) ProcessFile(ctx context.Context, path string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Skip .pml directory
	if strings.Contains(path, ".pml/") {
		return nil
	}

	// Check if path is a directory
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if info.IsDir() {
		return nil
	}

	// Read file content with UTF-8 encoding
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate file checksum for cache
	fileChecksum := p.calculateChecksum(string(content))

	// Parse blocks from content
	blocks, err := p.parseBlocks(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse blocks: %w", err)
	}

	// Create results directory if it doesn't exist
	resultsDir := filepath.Join(filepath.Dir(path), ".pml", "results")
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	// Initialize or update cache entry for the file
	p.cacheMu.Lock()
	entry, ok := p.cache[path]
	if !ok || entry.Checksum != fileChecksum {
		entry = CacheEntry{
			Checksum: fileChecksum,
			ModTime:  time.Now(),
			Blocks:   make(map[string]BlockCache),
		}
	}
	p.cache[path] = entry
	p.cacheMu.Unlock()

	// Process each block
	var wg sync.WaitGroup
	errChan := make(chan error, len(blocks))
	resultFiles := make([]string, len(blocks))
	var resultsMu sync.Mutex

	// Create a semaphore to limit concurrent goroutines
	semaphore := make(chan struct{}, 10) // Process up to 10 blocks concurrently

	// Process blocks in order to maintain consistent result file names
	for i := range blocks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Process block using processBlock function
				resultFile, err := p.processBlock(ctx, blocks[i], i, path, filepath.Dir(path))
				if err != nil {
					errChan <- fmt.Errorf("failed to process block %d: %w", i, err)
					return
				}

				// Store result file
				resultsMu.Lock()
				resultFiles[i] = resultFile
				resultsMu.Unlock()
			}(i)
		}
	}

	// Wait for all blocks to be processed
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(errChan)
	}()

	// Wait for completion or cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		// Check for errors
		var errs []error
		for err := range errChan {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("multiple errors: %v", errs)
		}
	}

	// Update content with results
	newContent := p.updateContentWithResults(blocks, string(content), resultFiles, resultsDir, filepath.Base(path))

	// Write updated content back to file with UTF-8 encoding
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated file: %w", err)
	}

	// Save cache to disk
	if err := p.saveCache(); err != nil {
		p.debugf("Warning: failed to save cache: %v\n", err)
	}

	return nil
}

// processBlock processes a single block and returns its result
func (p *Parser) processBlock(ctx context.Context, block Block, index int, plmPath string, localResultsDir string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Calculate block checksum for caching
	blockChecksum := p.calculateBlockChecksum(block)

	// Check cache for this block using checksum as key
	if !p.forceProcess {
		p.cacheMu.Lock()
		entry, ok := p.cache[plmPath]
		if ok {
			if blockCache, ok := entry.Blocks[blockChecksum]; ok {
				p.cacheMu.Unlock()
				return blockCache.Result, nil
			}
		}
		p.cacheMu.Unlock()
	}

	// Process the block based on its type
	var result string
	var err error
	switch block.Type {
	case DirectiveAsk, DirectiveDo:
		result, err = p.llm.Ask(ctx, strings.Join(block.Content, "\n"))
	default:
		return "", fmt.Errorf("unknown block type: %s", block.Type)
	}

	if err != nil {
		return "", fmt.Errorf("failed to process block: %w", err)
	}

	// Create results directory if it doesn't exist
	resultsDir := filepath.Join(localResultsDir, ".pml", "results")
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create results directory: %w", err)
	}

	// Generate a unique result file name
	resultFile := p.generateUniqueResultName(filepath.Base(plmPath), index, block.Type, resultsDir)

	// Create summary for the result
	summary := fmt.Sprintf("Result for block %d from %s", index, filepath.Base(plmPath))

	// Write the result to a file with proper format
	err = p.writeResult(block, result, resultFile, resultsDir, summary)
	if err != nil {
		return "", fmt.Errorf("failed to write result: %w", err)
	}

	// Update cache entry for this block
	p.cacheMu.Lock()
	entry, ok := p.cache[plmPath]
	if !ok {
		entry = CacheEntry{
			Blocks: make(map[string]BlockCache),
		}
	}
	entry.Blocks[blockChecksum] = BlockCache{
		Checksum: blockChecksum,
		Result:   result,
		ModTime:  time.Now(),
	}
	p.cache[plmPath] = entry
	p.cacheMu.Unlock()

	return resultFile, nil
}

// writeResult writes a block's result to a file
func (p *Parser) writeResult(block Block, result string, resultFile string, localResultsDir string, summary string) error {
	// Format the result with metadata and content
	metadata := map[string]interface{}{
		"is_ephemeral": true,
		"type":         block.Type,
		"summary":      summary,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Format the content with UTF-8 encoding preserved
	content := fmt.Sprintf("# metadata:%s\n\nQuestion:\n%s\n\nAnswer:\n%s\n",
		string(metadataJSON),
		strings.Join(block.Content, "\n"),
		result)

	// Write the result file with UTF-8 encoding
	resultPath := filepath.Join(localResultsDir, resultFile)
	err = os.WriteFile(resultPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write result file: %w", err)
	}

	return nil
}

// updateContentWithResults updates the original content by generating result files
// for each block and embedding a result link in place of the block.
func (p *Parser) updateContentWithResults(blocks []Block, content string, resultFiles []string, localResultsDir string, sourceFile string) string {
	if len(blocks) == 0 {
		return content
	}

	var newContent strings.Builder
	lastPos := 0

	for i, block := range blocks {
		// Write content before this block
		newContent.WriteString(content[lastPos:block.Start])

		// Insert a link in the original .pml
		// Include the full path relative to the source file
		relPath := resultFiles[i]
		if strings.HasPrefix(relPath, "--(r/") {
			relPath = strings.TrimPrefix(relPath, "--(r/")
			relPath = strings.TrimSuffix(relPath, ")")
		}
		if strings.HasPrefix(relPath, ":--(r/") {
			relPath = strings.TrimPrefix(relPath, ":--(r/")
			relPath = strings.TrimSuffix(relPath, ")")
		}
		newContent.WriteString(fmt.Sprintf(":--(r/%s)", relPath))

		lastPos = block.End
	}

	// Write anything after the last block
	if lastPos < len(content) {
		newContent.WriteString(content[lastPos:])
	}

	return newContent.String()
}
