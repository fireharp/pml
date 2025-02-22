package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ProcessFile processes a single PML file (parse, generate .py, run blocks in parallel)
func (p *Parser) ProcessFile(ctx context.Context, plmPath string) error {
    if ctx == nil {
        ctx = context.Background()
    }
    if err := ctx.Err(); err != nil {
        return err
    }

    // Get or create a file lock
    lockInterface, _ := p.fileLocks.LoadOrStore(plmPath, &sync.Mutex{})
    fileLock := lockInterface.(*sync.Mutex)
    fileLock.Lock()
    defer fileLock.Unlock()
	// Skip .pml/ directories and check if the path is a directory
	if strings.Contains(plmPath, "/.pml/") || strings.Contains(plmPath, "\\.pml\\") {
		return nil
	}

	var (
		fileInfo   os.FileInfo
		content    []byte
		err        error
		blocks     []Block
		newContent string
	)

	fileInfo, err = os.Stat(plmPath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot process directory as PML file: %s", plmPath)
	}

	content, err = os.ReadFile(plmPath)
	if err != nil {
		return fmt.Errorf("failed to read plm file: %w", err)
	}
	// (Removed) Do not skip processing even if a result link is present.

	// Use the parser’s designated results directory.
	resultsDir := p.rootResultsDir
	if err = os.MkdirAll(resultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	// Make sure rootResultsDir exists
	if err = os.MkdirAll(p.rootResultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create root results dir: %w", err)
	}
	if err = p.ensureDirectories(); err != nil {
		return err
	}

	// parse blocks
	blocks, err = p.parseBlocks(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse blocks: %w", err)
	}

	// create .py file
	newContent = p.replaceBlocksInContent(string(content), blocks)
	pyPath := filepath.Join(filepath.Dir(plmPath), filepath.Base(plmPath)+".py")
	if err = os.WriteFile(pyPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write Python file: %w", err)
	}

	// process blocks in parallel
	results := make([]string, len(blocks))
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error
	var resultMu sync.Mutex
	pmlDir := filepath.Dir(plmPath)
	for i := range blocks {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errMu.Lock()
				if firstErr == nil {
					firstErr = ctx.Err()
				}
				errMu.Unlock()
				return
			default:
				result, err := p.processBlock(ctx, blocks[i], i, plmPath, pmlDir)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = fmt.Errorf("failed to process block %d: %w", i, err)
					}
					errMu.Unlock()
					return
				}

				resultMu.Lock()
				results[i] = result
				resultMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	// Create results directory if it doesn't exist
	if err := os.MkdirAll(pmlDir, 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	// embed results
	updatedContent := p.updateContentWithResults(blocks, string(content), results, resultsDir, filepath.Base(plmPath))
	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(plmPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated PML file: %w", err)
	}

	// update cache using block checksums as keys
	fileInfo, err = os.Stat(plmPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	p.cacheMu.Lock()
	entry := p.cache[plmPath]
	entry.Checksum = p.calculateChecksum(string(content))
	entry.ModTime = fileInfo.ModTime()
	if entry.Blocks == nil {
		entry.Blocks = make(map[string]BlockCache)
	}
	for i, block := range blocks {
		blockChecksum := p.calculateBlockChecksum(block)
		entry.Blocks[blockChecksum] = BlockCache{
			Checksum: blockChecksum,
			Result:   results[i],
			ModTime:  time.Now(),
		}
	}
	p.cache[plmPath] = entry
	p.cacheMu.Unlock()

	// Save cache once at the end
	if err := p.saveCache(); err != nil {
		p.debugf("Warning: failed to save cache: %v\n", err)
	}

	return nil
}

// processBlock processes a single block and returns its result
func (p *Parser) processBlock(ctx context.Context, block Block, index int, plmPath string, pmlDir string) (string, error) {
    if err := ctx.Err(); err != nil {
        return "", err
    }
	blockChecksum := p.calculateBlockChecksum(block)

	// Check cache for this block using checksum as key
	if !p.forceProcess {
		p.cacheMu.Lock()
		entry, ok := p.cache[plmPath]
		if ok {
			if blockCache, ok := entry.Blocks[blockChecksum]; ok {
				p.cacheMu.Unlock()
				p.debugf("Cache hit for block %d in %s\n", index, plmPath)
				return blockCache.Result, nil
			}
		}
		p.cacheMu.Unlock()
	}

	var result string
	var err error

	switch block.Type {
	case DirectiveAsk, DirectiveDo:
		// Use the same LLM response for both types
		result, err = p.llm.Ask(ctx, strings.Join(block.Content, "\n"))
	default:
		return "", fmt.Errorf("unknown block type: %s", block.Type)
	}

	if err != nil {
		return "", fmt.Errorf("failed to process block: %w", err)
	}

	// Update cache entry for this block
	p.cacheMu.Lock()
	entry, ok := p.cache[plmPath]
	if !ok {
		// Create new cache entry if it doesn't exist
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

	return result, nil
}

// writeResult writes a block's result to a file
func (p *Parser) writeResult(block Block, result string, resultFile string, localResultsDir string, summary string) error {
	// Ensure the local results directory exists
	if err := os.MkdirAll(localResultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create local results directory: %w", err)
	}

	resultPath := filepath.Join(localResultsDir, resultFile)

	// Check if another goroutine is already writing this file
	if _, exists := p.resultFiles.LoadOrStore(resultPath, true); exists {
		// Another goroutine is writing this file, wait a bit and check if it exists
		time.Sleep(10 * time.Millisecond)
		if _, err := os.Stat(resultPath); err == nil {
			// File exists, we can use it
			p.resultFiles.Delete(resultPath)
			return nil
		}
		// File still doesn't exist, proceed with writing
	}
	defer p.resultFiles.Delete(resultPath)

	// Format the result content with proper escaping
	content := fmt.Sprintf("// %s\n\nQuestion:\n%s\n\nAnswer:\n%s\n", 
		summary, 
		strings.Join(block.Content, "\n"),
		result)

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(resultPath), 0755); err != nil {
		return fmt.Errorf("failed to create result directory: %w", err)
	}

	// Write the result file
	if err := os.WriteFile(resultPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write result file: %w", err)
	}

	p.debugf("Wrote result to %s\n", resultPath)
	return nil
}

// updateContentWithResults updates the original content by generating result files
// for each block and embedding a result link in place of the block.
func (p *Parser) updateContentWithResults(blocks []Block, content string, results []string, localResultsDir string, sourceFile string) string {
	if len(blocks) == 0 {
		return content
	}

	var newContent strings.Builder
	lastPos := 0

	for i, block := range blocks {
		// Write content before this block
		newContent.WriteString(content[lastPos:block.Start])

		// Generate unique result file name
		uniqueName := p.generateUniqueResultName(sourceFile, i, block.Type, localResultsDir)
		resultFile := uniqueName + ".pml"
		summary := fmt.Sprintf("Result for block %d from %s", i, sourceFile)

		// Create ephemeral content with metadata, question and answer
		questionText := strings.Join(block.Content, "\n")
		ephemeralContent := fmt.Sprintf(`# metadata:{"is_ephemeral":true}

Question:
%s

Answer:
%s
`, questionText, results[i])

		if err := p.writeResult(block, ephemeralContent, resultFile, localResultsDir, summary); err != nil {
			// If write fails, just inline the result as fallback
			newContent.WriteString(results[i])
		} else {
			// Insert a link in the original .pml
			newContent.WriteString(fmt.Sprintf(":--(r/%s:\"%s\")", uniqueName, results[i]))
		}

		lastPos = block.End
	}

	// Write anything after the last block
	if lastPos < len(content) {
		newContent.WriteString(content[lastPos:])
	}

	return newContent.String()
}
