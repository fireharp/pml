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
	// Skip .pml/ directories
	if strings.Contains(plmPath, "/.pml/") || strings.Contains(plmPath, "\\.pml\\") {
		return nil
	}

	content, err := os.ReadFile(plmPath)
	if err != nil {
		return fmt.Errorf("failed to read plm file: %w", err)
	}

	// Prepare .pml directory
	pmlDir := filepath.Join(filepath.Dir(plmPath), ".pml")
	if err := os.MkdirAll(pmlDir, 0755); err != nil {
		return fmt.Errorf("failed to create .pml dir: %w", err)
	}

	// Make sure rootResultsDir exists
	if err := os.MkdirAll(p.rootResultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create root results dir: %w", err)
	}
	if err := p.ensureDirectories(); err != nil {
		return err
	}

	// parse blocks
	blocks, err := p.parseBlocks(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse blocks: %w", err)
	}

	// create .py file
	newContent := p.replaceBlocksInContent(string(content), blocks)
	pyPath := plmPath + ".py"
	if err := os.WriteFile(pyPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write Python file: %w", err)
	}

	// process blocks in parallel
	results := make([]string, len(blocks))
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error
	var resultMu sync.Mutex

	for i := range blocks {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

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
		}(i)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	// embed results
	updatedContent := p.updateContentWithResults(blocks, string(content), results, pmlDir, filepath.Base(plmPath))
	if err := os.WriteFile(plmPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated PML file: %w", err)
	}

	// update cache using block checksums as keys
	fileInfo, err := os.Stat(plmPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

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
	if err := p.saveCache(); err != nil {
		p.debugf("Warning: failed to save cache: %v\n", err)
	}

	return nil
}

// processBlock processes a single block and returns its result
func (p *Parser) processBlock(ctx context.Context, block Block, index int, plmPath string, pmlDir string) (string, error) {
	blockChecksum := p.calculateBlockChecksum(block)

	// Check cache for this block using checksum as key
	if !p.forceProcess {
		if entry, ok := p.cache[plmPath]; ok {
			if blockCache, ok := entry.Blocks[blockChecksum]; ok {
				p.debugf("Cache hit for block %d in %s\n", index, plmPath)
				return blockCache.Result, nil
			}
		}
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

	// Update cache immediately after processing
	if entry, ok := p.cache[plmPath]; ok {
		entry.Blocks[blockChecksum] = BlockCache{
			Checksum: blockChecksum,
			Result:   result,
			ModTime:  time.Now(),
		}
		p.cache[plmPath] = entry
	} else {
		// Create new cache entry if it doesn't exist
		p.cache[plmPath] = CacheEntry{
			Blocks: map[string]BlockCache{
				blockChecksum: {
					Checksum: blockChecksum,
					Result:   result,
					ModTime:  time.Now(),
				},
			},
		}
	}

	// Save cache to disk
	if err := p.saveCache(); err != nil {
		p.debugf("Warning: failed to save cache: %v\n", err)
	}

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
			return nil
		}
		// File still doesn't exist, proceed with writing
	}
	defer p.resultFiles.Delete(resultPath)

	// Format the result content with proper escaping
	content := fmt.Sprintf("// %s\n\n%s\n", summary, result)

	// Write the result file
	if err := os.WriteFile(resultPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write result file: %w", err)
	}

	p.debugf("Wrote result to %s\n", resultPath)
	return nil
}

// updateContentWithResults updates the original content with result links
func (p *Parser) updateContentWithResults(blocks []Block, content string, results []string, localResultsDir string, sourceFile string) string {
	lines := strings.Split(content, "\n")
	var output []string
	var currentBlock int

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		switch {
		case trimmedLine == DirectiveAsk || trimmedLine == DirectiveDo:
			output = append(output, line)
		case trimmedLine == DirectiveEnd:
			if currentBlock < len(blocks) && len(results) > currentBlock {
				resultName := p.generateUniqueResultName(sourceFile, currentBlock, localResultsDir)
				resultPath := filepath.Join(localResultsDir, fmt.Sprintf("%s.pml", resultName))

				// Write the result file if it doesn't exist
				if _, err := os.Stat(resultPath); os.IsNotExist(err) {
					summary := fmt.Sprintf("Result for block %d from %s", currentBlock, sourceFile)
					if err := p.writeResult(blocks[currentBlock], results[currentBlock], fmt.Sprintf("%s.pml", resultName), localResultsDir, summary); err != nil {
						p.debugf("Failed to write result file: %v\n", err)
						output = append(output, results[currentBlock])
						currentBlock++
						continue
					}
				}

				// Add the result link
				output = append(output, fmt.Sprintf(":--(r/%s:\"%s\")", resultName, results[currentBlock]))
			} else {
				output = append(output, line)
			}
			currentBlock++
		default:
			output = append(output, line)
		}
	}

	return strings.Join(output, "\n")
}
