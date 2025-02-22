package parser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DirectiveAsk = ":ask"
	DirectiveDo  = ":do"
	DirectiveEnd = ":--"
)

// LLMClient interface for making LLM requests
type LLMClient interface {
	Ask(ctx context.Context, prompt string) (string, error)
	Summarize(ctx context.Context, text string) (string, error)
}

// CacheEntry represents a cached processing result
type CacheEntry struct {
	Checksum string                `json:"checksum"`
	ModTime  time.Time             `json:"mod_time"`
	Blocks   map[string]BlockCache `json:"blocks"`
}

// BlockCache represents a cached block processing result
type BlockCache struct {
	Checksum string    `json:"checksum"`
	Result   string    `json:"result"`
	ModTime  time.Time `json:"mod_time"`
}

// Parser handles PML file parsing and processing
type Parser struct {
	llm            LLMClient
	sourcesDir     string
	compiledDir    string
	rootResultsDir string // For larger logs and detailed execution results
	cacheFile      string // Path to the cache file
	cache          map[string]CacheEntry
	debug          bool
	forceProcess   bool
}

// FileBlocks holds the original file path plus the parsed blocks
type FileBlocks struct {
	FilePath string
	Blocks   []Block
}

// BlockResult holds the final result for a single block
type BlockResult struct {
	FilePath string
	BlockIdx int
	Block    Block
	Result   string
	Err      error
}

// NewParser creates a new PML parser with specified directories
func NewParser(llm LLMClient, sourcesDir, compiledDir, resultsDir string) *Parser {
	// Cache file is now stored in the .pml directory
	pmlDir := filepath.Join(sourcesDir, ".pml")
	cacheFile := filepath.Join(pmlDir, "cache.json")
	p := &Parser{
		llm:            llm,
		sourcesDir:     sourcesDir,
		compiledDir:    compiledDir, // Keep for compatibility, but will be same as sourcesDir
		rootResultsDir: resultsDir,
		cacheFile:      cacheFile,
		cache:          make(map[string]CacheEntry),
		debug:          os.Getenv("PML_DEBUG") == "1",
		forceProcess:   false,
	}
	p.loadCache()
	return p
}

// loadCache loads the cache from disk
func (p *Parser) loadCache() {
	data, err := os.ReadFile(p.cacheFile)
	if err != nil {
		p.debugf("No cache file found or error reading cache: %v\n", err)
		return
	}

	if err := json.Unmarshal(data, &p.cache); err != nil {
		p.debugf("Error unmarshaling cache: %v\n", err)
		// Start with empty cache if corrupted
		p.cache = make(map[string]CacheEntry)
	}
}

// saveCache saves the cache to disk
func (p *Parser) saveCache() error {
	data, err := json.MarshalIndent(p.cache, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling cache: %w", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(p.cacheFile), 0755); err != nil {
		return fmt.Errorf("error creating cache directory: %w", err)
	}

	if err := os.WriteFile(p.cacheFile, data, 0644); err != nil {
		return fmt.Errorf("error writing cache file: %w", err)
	}

	p.debugf("Cache saved to %s\n", p.cacheFile)
	return nil
}

// calculateChecksum calculates SHA-256 checksum of file content, ignoring result links
func (p *Parser) calculateChecksum(content string) string {
	// Remove result links before calculating checksum
	resultLinkPattern := regexp.MustCompile(`:-+\(r/[a-z]+_[a-z]+\)`)
	contentWithoutLinks := resultLinkPattern.ReplaceAllString(content, ":--")

	hash := sha256.Sum256([]byte(contentWithoutLinks))
	return hex.EncodeToString(hash[:])
}

// debugf prints debug messages if debug mode is enabled
func (p *Parser) debugf(format string, args ...interface{}) {
	if p.debug {
		fmt.Printf(format, args...)
	}
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

// Block represents a block in PML file
type Block struct {
	Type        string
	Content     []string
	Response    string
	IsEphemeral bool // Whether this block was generated during runtime
}

var (
	adjectives = []string{
		"happy", "clever", "swift", "gentle", "brave",
		"wise", "calm", "bright", "kind", "quick",
		"silent", "proud", "bold", "eager", "peaceful",
		"witty", "warm", "smart", "noble", "merry",
	}

	nouns = []string{
		"panda", "falcon", "dolphin", "phoenix", "tiger",
		"maple", "river", "mountain", "breeze", "cloud",
		"crystal", "garden", "meadow", "ocean", "forest",
		"sunrise", "comet", "rainbow", "valley", "whisper",
	}
)

// generateUniqueResultName generates a friendly name for a result file that is guaranteed to be unique
func (p *Parser) generateUniqueResultName(sourceFile string, blockIndex int, localResultsDir string) string {
	counter := 0
	var resultName string
	for {
		// Generate base name using file hash and block index
		hash := 0
		for _, c := range sourceFile {
			hash = (hash*31 + int(c)) % len(nouns)
		}

		// Use blockIndex, counter and file hash to ensure uniqueness
		adjIndex := (blockIndex + hash + counter) % len(adjectives)
		nounIndex := ((blockIndex + hash + counter) * 7) % len(nouns)
		resultName = fmt.Sprintf("%s_%s", adjectives[adjIndex], nouns[nounIndex])

		// Check if this name is already taken
		resultFile := fmt.Sprintf("%s.pml", resultName)
		resultPath := filepath.Join(localResultsDir, resultFile)
		if _, err := os.Stat(resultPath); os.IsNotExist(err) {
			// Name is available
			return resultName
		}
		// Name is taken, try next combination
		counter++

		// Safety check to prevent infinite loop (though extremely unlikely)
		if counter > len(adjectives)*len(nouns) {
			// If we somehow exhausted all combinations, use timestamp as fallback
			return fmt.Sprintf("result_%d", time.Now().UnixNano())
		}
	}
}

// SetForceProcess sets whether to force process files regardless of cache
func (p *Parser) SetForceProcess(force bool) {
	p.forceProcess = force
}

// calculateBlockChecksum calculates SHA-256 checksum of a block's content, ignoring whitespace
func (p *Parser) calculateBlockChecksum(block Block) string {
	// Normalize block content by trimming whitespace
	var normalized strings.Builder
	normalized.WriteString(strings.TrimSpace(block.Type))
	for _, line := range block.Content {
		normalized.WriteString(strings.TrimSpace(line))
	}

	hash := sha256.Sum256([]byte(normalized.String()))
	return hex.EncodeToString(hash[:])
}

// ProcessFile processes a PML file, handling any blocks and compiling to Python
func (p *Parser) ProcessFile(ctx context.Context, plmPath string) error {
	// Skip files in .pml/ directory
	if strings.Contains(plmPath, "/.pml/") || strings.Contains(plmPath, "\\.pml\\") {
		return nil
	}

	// Read the PML file
	content, err := os.ReadFile(plmPath)
	if err != nil {
		return fmt.Errorf("failed to read plm file: %w", err)
	}

	// Calculate checksum
	checksum := p.calculateChecksum(string(content))

	// Get file modification time for logging only
	fileInfo, err := os.Stat(plmPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	modTime := fileInfo.ModTime()

	// Initialize or update cache entry
	if entry, ok := p.cache[plmPath]; !ok || entry.Blocks == nil {
		p.cache[plmPath] = CacheEntry{
			Checksum: checksum,
			ModTime:  modTime,
			Blocks:   make(map[string]BlockCache),
		}
	}

	// Create .pml directory next to the PML file for local results
	pmlDir := filepath.Join(filepath.Dir(plmPath), ".pml")
	localResultsDir := pmlDir
	blocksDir := filepath.Join(pmlDir, "blocks")
	if err := os.MkdirAll(blocksDir, 0755); err != nil {
		return fmt.Errorf("failed to create blocks directory: %w", err)
	}

	// Create root results directory for detailed logs
	if err := os.MkdirAll(p.rootResultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create root results directory: %w", err)
	}

	if err := p.ensureDirectories(); err != nil {
		return err
	}

	// Parse blocks from the PML file
	blocks, err := p.parseBlocks(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse blocks: %w", err)
	}

	// Create exact copy with replaced blocks
	newContent := p.replaceBlocksInContent(string(content), blocks)
	// Put the main Python file in the same directory as the source file
	pyPath := plmPath + ".py"
	if err := os.WriteFile(pyPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write Python file: %w", err)
	}

	// Process each block separately and collect results
	var results []string
	for i, block := range blocks {
		result, err := p.processBlock(ctx, block, i, plmPath, blocksDir)
		if err != nil {
			return fmt.Errorf("failed to process block %d: %w", i, err)
		}
		results = append(results, result)
	}

	// Update PML file with results
	newContent = p.updateContentWithResults(blocks, string(content), results, localResultsDir, filepath.Base(plmPath))
	if err := os.WriteFile(plmPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated PML file: %w", err)
	}

	// Update cache after successful processing
	entry := p.cache[plmPath]
	entry.Checksum = checksum
	entry.ModTime = modTime
	for i, block := range blocks {
		blockChecksum := p.calculateBlockChecksum(block)
		blockKey := fmt.Sprintf("%s_block_%d", filepath.Base(plmPath), i)
		entry.Blocks[blockKey] = BlockCache{
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

func (p *Parser) parseBlocks(content string) ([]Block, error) {
	lines := strings.Split(content, "\n")
	var blocks []Block
	var currentBlock *Block
	var blockContent []string

	p.debugf("Parsing blocks from content:\n%s\n", content)

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		p.debugf("Processing line: %q\n", trim)

		switch {
		case strings.HasPrefix(trim, DirectiveAsk), strings.HasPrefix(trim, DirectiveDo):
			// If we have an open block, close it before starting a new one
			if currentBlock != nil {
				currentBlock.Content = blockContent
				blocks = append(blocks, *currentBlock)
				p.debugf("Found new directive, closing block with content: %q\n", strings.Join(blockContent, "\n"))
			}
			p.debugf("Found directive: %s\n", trim)
			currentBlock = &Block{Type: trim}
			blockContent = []string{}
		case strings.HasPrefix(trim, DirectiveEnd), strings.HasPrefix(trim, ":--(r/"):
			if currentBlock != nil {
				currentBlock.Content = blockContent
				p.debugf("Found end of block with content: %q\n", strings.Join(blockContent, "\n"))
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
				blockContent = nil
			}
		case strings.HasPrefix(trim, ":"):
			// Any other directive starts a new block
			if currentBlock != nil {
				currentBlock.Content = blockContent
				p.debugf("Found new directive, closing block with content: %q\n", strings.Join(blockContent, "\n"))
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
				blockContent = nil
			}
		default:
			if currentBlock != nil {
				// If this is a blank line followed by a blank line or end of file,
				// or if this is the last line, close the current block
				if trim == "" {
					if i == len(lines)-1 || (i < len(lines)-1 && strings.TrimSpace(lines[i+1]) == "") {
						currentBlock.Content = blockContent
						p.debugf("Found implicit block end with content: %q\n", strings.Join(blockContent, "\n"))
						blocks = append(blocks, *currentBlock)
						currentBlock = nil
						blockContent = nil
						continue
					}
				}
				blockContent = append(blockContent, line)
			}
		}
	}

	// If we have an open block at the end of file, close it
	if currentBlock != nil {
		currentBlock.Content = blockContent
		p.debugf("Found end of file with open block: %q\n", strings.Join(blockContent, "\n"))
		blocks = append(blocks, *currentBlock)
	}

	p.debugf("Found %d blocks\n", len(blocks))
	for i, block := range blocks {
		p.debugf("Block %d: Type=%s, Content=%q\n", i, block.Type, strings.Join(block.Content, "\n"))
	}

	return blocks, nil
}

// replaceBlocksInContent creates a copy of the original content with blocks replaced by Python
func (p *Parser) replaceBlocksInContent(content string, blocks []Block) string {
	lines := strings.Split(content, "\n")
	var newLines []string
	var currentBlock *Block
	blockIndex := 0
	resultNames := p.extractResultNames(content)

	for _, line := range lines {
		trim := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trim, DirectiveAsk), strings.HasPrefix(trim, DirectiveDo):
			currentBlock = &Block{Type: trim}
			// Add Python equivalent with comment showing original directive
			newLines = append(newLines, fmt.Sprintf("# %s", line))
			resultVar := fmt.Sprintf("result_%d", blockIndex)
			if strings.HasPrefix(trim, DirectiveAsk) {
				newLines = append(newLines, fmt.Sprintf("%s = process_ask('''", resultVar))
			} else {
				newLines = append(newLines, fmt.Sprintf("%s = process_do('''", resultVar))
			}
		case strings.HasPrefix(trim, DirectiveEnd), strings.HasPrefix(trim, ":--(r/"):
			if currentBlock != nil {
				newLines = append(newLines, "''')")
				newLines = append(newLines, "# :--")
				// If we have a result name for this block, use it in the comment
				if blockIndex < len(resultNames) && resultNames[blockIndex] != "" {
					newLines = append(newLines, fmt.Sprintf("# :--(r/%s)", resultNames[blockIndex]))
				}
				blockIndex++
				currentBlock = nil
			}
		case strings.HasPrefix(trim, ":"):
			// Any other directive starts a new block
			if currentBlock != nil {
				newLines = append(newLines, "''')")
				newLines = append(newLines, "# :--")
				blockIndex++
				currentBlock = nil
			}
			newLines = append(newLines, line)
		default:
			if currentBlock != nil {
				newLines = append(newLines, line)
			} else {
				newLines = append(newLines, line)
			}
		}
	}

	// Close any open block at the end of file
	if currentBlock != nil {
		newLines = append(newLines, "''')")
		newLines = append(newLines, "# :--")
		blockIndex++
	}

	// Add necessary imports at the top if there were any blocks
	if blockIndex > 0 {
		imports := []string{
			"# Auto-generated imports for PML blocks",
			"from src.pml.directives import process_ask, process_do",
			"",
		}
		newLines = append(imports, newLines...)
	}

	return strings.Join(newLines, "\n")
}

// extractResultNames extracts existing result names from PML file
func (p *Parser) extractResultNames(content string) []string {
	var resultNames []string
	lines := strings.Split(content, "\n")
	resultLinkPattern := regexp.MustCompile(`:-+\(r/([^)]+)\)`)

	for _, line := range lines {
		if matches := resultLinkPattern.FindStringSubmatch(strings.TrimSpace(line)); len(matches) > 1 {
			resultNames = append(resultNames, matches[1])
		}
	}

	return resultNames
}

// getBlocksDir returns the path to the .pml/blocks directory and ensures it exists
func (p *Parser) getBlocksDir(plmPath string) (string, error) {
	pmlDir := filepath.Join(filepath.Dir(plmPath), ".pml")
	blocksDir := filepath.Join(pmlDir, "blocks")
	if err := os.MkdirAll(blocksDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create blocks directory: %w", err)
	}
	return blocksDir, nil
}

// processBlock processes a single block and returns its result
func (p *Parser) processBlock(ctx context.Context, block Block, index int, plmPath string, blocksDir string) (string, error) {
	// Calculate block checksum
	blockChecksum := p.calculateBlockChecksum(block)
	blockKey := fmt.Sprintf("%s_block_%d", filepath.Base(plmPath), index)

	// Check cache for this block
	if entry, ok := p.cache[plmPath]; ok && !p.forceProcess {
		if blockCache, ok := entry.Blocks[blockKey]; ok {
			if blockCache.Checksum == blockChecksum {
				p.debugf("Block %d unchanged, using cached result\n", index)
				return blockCache.Result, nil
			}
		}
	}

	// Check if we're in test mode
	testDirPath := filepath.Join(filepath.Dir(plmPath), "test_directives.py")
	if _, err := os.Stat(testDirPath); err == nil {
		// In test mode, use LLM directly
		content := strings.Join(block.Content, "\n")
		result, err := p.llm.Ask(ctx, content)
		if err != nil {
			return "", fmt.Errorf("failed to process block with LLM: %w", err)
		}

		// Update cache with block result
		if entry, ok := p.cache[plmPath]; ok {
			if entry.Blocks == nil {
				entry.Blocks = make(map[string]BlockCache)
			}
			entry.Blocks[blockKey] = BlockCache{
				Checksum: blockChecksum,
				Result:   result,
				ModTime:  time.Now(),
			}
			p.cache[plmPath] = entry
		}

		return result, nil
	}

	// Create a block file in the blocks directory with a unique name
	blockFileName := fmt.Sprintf(".%s.block_%d.py", filepath.Base(plmPath), index)
	blockFile := filepath.Join(blocksDir, blockFileName)

	var code strings.Builder
	code.WriteString("# Auto-generated from PML block\n")
	code.WriteString("import os\n")
	code.WriteString("import sys\n\n")

	// Add impl1 directory to Python path
	code.WriteString("# Add impl1 directory to Python path\n")
	code.WriteString("current_dir = os.path.dirname(os.path.abspath(__file__))\n")
	code.WriteString("impl1_dir = os.path.abspath(os.path.join(current_dir, '..', '..', '..'))\n")
	code.WriteString("src_dir = os.path.join(impl1_dir, 'src')\n")
	code.WriteString("if impl1_dir not in sys.path:\n")
	code.WriteString("    sys.path.insert(0, impl1_dir)\n")
	code.WriteString("if src_dir not in sys.path:\n")
	code.WriteString("    sys.path.insert(0, src_dir)\n\n")

	// Add imports based on test mode
	code.WriteString("from src.pml.directives import process_ask, process_do\n")
	code.WriteString("\n")

	// Add the block processing code
	resultVar := fmt.Sprintf("result_%d", index)
	if block.Type == DirectiveAsk {
		code.WriteString(fmt.Sprintf("%s = process_ask('''\n", resultVar))
	} else {
		code.WriteString(fmt.Sprintf("%s = process_do('''\n", resultVar))
	}
	code.WriteString(strings.Join(block.Content, "\n"))
	code.WriteString("\n''')\n")
	code.WriteString(fmt.Sprintf("print(%s)\n", resultVar))

	if err := os.WriteFile(blockFile, []byte(code.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write block file: %w", err)
	}

	// Execute the block file with proper Python path
	result, err := p.executePython(ctx, blockFile)
	if err != nil {
		return "", err
	}

	// Keep the block file for reference (don't remove it)

	// Update cache with block result
	if entry, ok := p.cache[plmPath]; ok {
		if entry.Blocks == nil {
			entry.Blocks = make(map[string]BlockCache)
		}
		entry.Blocks[blockKey] = BlockCache{
			Checksum: blockChecksum,
			Result:   result[0],
			ModTime:  time.Now(),
		}
		p.cache[plmPath] = entry
	}

	if len(result) > 0 {
		return result[0], nil
	}
	return "", nil
}

// updateContentWithResults updates the PML content with links to results
func (p *Parser) updateContentWithResults(blocks []Block, content string, results []string, localResultsDir string, sourceFile string) string {
	lines := strings.Split(content, "\n")
	var newLines []string
	var currentBlock *Block
	blockIndex := 0

	for _, line := range lines {
		trim := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trim, DirectiveAsk), strings.HasPrefix(trim, DirectiveDo):
			currentBlock = &Block{Type: trim}
			newLines = append(newLines, line)
		case strings.HasPrefix(trim, DirectiveEnd):
			if currentBlock != nil && blockIndex < len(results) {
				// Generate unique friendly name for this result
				resultName := p.generateUniqueResultName(sourceFile, blockIndex, localResultsDir)
				// Create result file in the local .pml directory
				resultFile := fmt.Sprintf("%s.pml", resultName)
				resultPath := filepath.Join(localResultsDir, resultFile)

				if err := p.writeResult(blocks[blockIndex], results[blockIndex], resultPath, localResultsDir, results[blockIndex]); err != nil {
					// If we can't write the result, just add it directly (fallback)
					newLines = append(newLines, line)
					newLines = append(newLines, results[blockIndex])
				} else {
					// Add link to result file with summary in the original format
					newLines = append(newLines, fmt.Sprintf(":--(r/%s:\"%s\")", resultName, results[blockIndex]))
				}
				blockIndex++
				currentBlock = nil
			} else {
				newLines = append(newLines, line)
			}
		default:
			newLines = append(newLines, line)
		}
	}

	return strings.Join(newLines, "\n")
}

// writeResult writes a result to a PML file with metadata
func (p *Parser) writeResult(block Block, result string, resultFile string, localResultsDir string, summary string) error {
	p.debugf("\n=== Writing result to file: %s ===\n", resultFile)
	p.debugf("Block type: %s\n", block.Type)
	p.debugf("Block content: %q\n", strings.Join(block.Content, "\n"))
	p.debugf("Result: %q\n", result)
	p.debugf("Summary: %q\n", summary)

	// Write detailed execution log to root results directory
	logName := fmt.Sprintf("%d_%s.log", time.Now().UnixNano(), filepath.Base(resultFile))
	logPath := filepath.Join(p.rootResultsDir, logName)
	p.debugf("Writing detailed log to: %s\n", logPath)
	detailedLog := fmt.Sprintf("Execution Time: %s\nProcess PID: %d\nBlock Type: %s\nContent:\n%s\n\nResult:\n%s\n",
		time.Now().Format(time.RFC3339),
		os.Getpid(),
		block.Type,
		strings.Join(block.Content, "\n"),
		result)
	p.debugf("Log content:\n%s\n", detailedLog)
	if err := os.WriteFile(logPath, []byte(detailedLog), 0644); err != nil {
		p.debugf("Warning: failed to write detailed log: %v\n", err)
		// Continue anyway as this is not critical
	}

	// Format the result properly
	formattedResult := p.formatResult(result)

	// Write PML result file with proper directive and summary
	var content strings.Builder
	content.WriteString(block.Type + "\n") // :ask or :do
	content.WriteString(strings.Join(block.Content, "\n"))
	content.WriteString("\n:--\n")
	content.WriteString("summary=")
	content.WriteString(p.formatString(summary))
	content.WriteString("\nresult=")
	content.WriteString(formattedResult)
	content.WriteString("\n")

	return os.WriteFile(resultFile, []byte(content.String()), 0644)
}

// formatResult formats a result value as valid PML
func (p *Parser) formatResult(result string) string {
	// If it looks like a number, boolean, or null, keep it as is
	if p.isLiteral(result) {
		return result
	}

	// Otherwise treat as string and properly escape
	return p.formatString(result)
}

// isLiteral checks if a string represents a literal value (number, boolean, null)
func (p *Parser) isLiteral(s string) bool {
	s = strings.TrimSpace(s)

	// Check for boolean and null
	if s == "true" || s == "false" || s == "null" {
		return true
	}

	// Check if it's a number (integer or float)
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}

	return false
}

// formatString properly escapes and quotes a string value
func (p *Parser) formatString(s string) string {
	// First, escape any special characters
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")

	// Wrap in quotes
	return fmt.Sprintf("\"%s\"", escaped)
}

// IsPMLFile checks if a file is a PML file
func IsPMLFile(path string) bool {
	// Skip files in .pml/ directory
	if strings.Contains(path, "/.pml/") || strings.Contains(path, "\\.pml\\") {
		return false
	}
	return strings.HasSuffix(strings.ToLower(path), ".pml")
}

// IsEphemeral checks if a PML file is an ephemeral block
func IsEphemeral(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# metadata:") {
			var metadata map[string]interface{}
			jsonStr := strings.TrimPrefix(line, "# metadata:")
			if err := json.Unmarshal([]byte(jsonStr), &metadata); err != nil {
				return false, err
			}
			if isEphemeral, ok := metadata["is_ephemeral"].(bool); ok {
				return isEphemeral, nil
			}
			break
		}
	}
	return false, nil
}

// ListEphemeralBlocks lists all ephemeral blocks in the results directory
func (p *Parser) ListEphemeralBlocks() ([]string, error) {
	var ephemeralBlocks []string

	files, err := os.ReadDir(p.rootResultsDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".pml") {
			path := filepath.Join(p.rootResultsDir, file.Name())
			isEphemeral, err := IsEphemeral(path)
			if err != nil {
				continue // Skip files with errors
			}
			if isEphemeral {
				ephemeralBlocks = append(ephemeralBlocks, path)
			}
		}
	}

	return ephemeralBlocks, nil
}

// executePython executes the compiled Python code and returns results
func (p *Parser) executePython(ctx context.Context, pyPath string) ([]string, error) {
	// Get project root directory (where impl1 directory is)
	projectRoot := filepath.Dir(filepath.Dir(p.sourcesDir)) // Go up two levels to get project root

	// Add both impl1 and src directories to PYTHONPATH
	env := os.Environ()
	pythonPathSet := false
	impl1Dir := filepath.Join(projectRoot, "impl1")
	srcDir := filepath.Join(projectRoot, "src")

	for i, e := range env {
		if strings.HasPrefix(e, "PYTHONPATH=") {
			// Append both impl1 and src directories to existing PYTHONPATH
			env[i] = e + string(os.PathListSeparator) + impl1Dir + string(os.PathListSeparator) + srcDir
			pythonPathSet = true
			break
		}
	}
	if !pythonPathSet {
		env = append(env, fmt.Sprintf("PYTHONPATH=%s%s%s", impl1Dir, string(os.PathListSeparator), srcDir))
	}

	// Use venv Python if it exists, otherwise use system Python
	venvPython := filepath.Join(projectRoot, ".venv", "bin", "python")
	python := "python"
	if _, err := os.Stat(venvPython); err == nil {
		python = venvPython
	}

	if p.debug {
		p.debugf("Executing Python with:\n")
		p.debugf("  Path: %s\n", pyPath)
		p.debugf("  Python: %s\n", python)
		p.debugf("  Project Root: %s\n", projectRoot)
		p.debugf("  Impl1 Dir: %s\n", impl1Dir)
		p.debugf("  Src Dir: %s\n", srcDir)
		for _, e := range env {
			if strings.HasPrefix(e, "PYTHONPATH=") {
				p.debugf("  %s\n", e)
			}
		}
	}

	cmd := exec.CommandContext(ctx, python, pyPath)
	cmd.Env = env

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute Python: %w\nOutput: %s", err, string(output))
	}

	// Split output into lines and return
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
}

// ProcessAllFiles concurrently processes all .pml files.
// 1) Parse + generate Python (fast)
// 2) Execute all blocks in parallel
// 3) Update final PML files
func (p *Parser) ProcessAllFiles(ctx context.Context) error {
	// 1) Find .pml files
	files, err := p.findPMLFiles()
	if err != nil {
		return err
	}

	// 2) Parse & compile stage
	// We can do this in parallel as well, but to keep it simple, do it sequentially
	var fileBlocks []FileBlocks
	for _, f := range files {
		fb, err := p.parseAndGeneratePython(f)
		if err != nil {
			return fmt.Errorf("parseAndGeneratePython failed for %s: %w", f, err)
		}
		fileBlocks = append(fileBlocks, fb)
	}

	// 3) Process all blocks in parallel
	// We'll spawn a goroutine for each block across all files.
	resultsCh := make(chan BlockResult)
	var wg sync.WaitGroup

	// Create a semaphore to limit concurrent goroutines
	const maxConcurrent = 10
	sem := make(chan struct{}, maxConcurrent)

	for _, fb := range fileBlocks {
		for i, blk := range fb.Blocks {
			wg.Add(1)
			go func(filePath string, idx int, block Block) {
				defer wg.Done()
				// Acquire semaphore
				sem <- struct{}{}
				defer func() { <-sem }()

				// Call processBlock, but it won't rewrite the PML file;
				// it just returns the result for us to store.
				result, err := p.processBlock(ctx, block, idx, filePath, filepath.Join(filepath.Dir(filePath), ".pml", "blocks"))
				resultsCh <- BlockResult{
					FilePath: filePath,
					BlockIdx: idx,
					Block:    block,
					Result:   result,
					Err:      err,
				}
			}(fb.FilePath, i, blk)
		}
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	blockResultsMap := make(map[string]map[int]BlockResult)
	for br := range resultsCh {
		if br.Err != nil {
			p.debugf("Error while processing block %d in %s: %v\n", br.BlockIdx, br.FilePath, br.Err)
			// Continue processing other blocks
		}

		// Group them by file -> blockIndex
		if _, ok := blockResultsMap[br.FilePath]; !ok {
			blockResultsMap[br.FilePath] = make(map[int]BlockResult)
		}
		blockResultsMap[br.FilePath][br.BlockIdx] = br
	}

	// 4) Finalize each .pml file with the results
	for _, fb := range fileBlocks {
		// Get all block results for this file
		results := make([]string, len(fb.Blocks))
		hasError := false
		for i := range fb.Blocks {
			if br, ok := blockResultsMap[fb.FilePath][i]; ok {
				if br.Err != nil {
					hasError = true
					results[i] = fmt.Sprintf("Error: %v", br.Err)
				} else {
					results[i] = br.Result
				}
			}
		}

		if hasError {
			p.debugf("Some blocks failed for %s, but continuing with available results\n", fb.FilePath)
		}

		// Read the original file content
		content, err := os.ReadFile(fb.FilePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", fb.FilePath, err)
		}

		// Update the file with all available results
		newContent := p.updatePMLFileWithResults(fb.Blocks, string(content), results, filepath.Join(filepath.Dir(fb.FilePath), ".pml"), filepath.Base(fb.FilePath))
		if err := os.WriteFile(fb.FilePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write updated PML file %s: %v", fb.FilePath, err)
		}
	}

	return nil
}

// findPMLFiles enumerates all .pml files in p.sourcesDir (recursive or not)
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

// parseAndGeneratePython parses a single .pml file to produce a list of blocks
// and writes out the Python code (no block processing/calls).
func (p *Parser) parseAndGeneratePython(plmPath string) (FileBlocks, error) {
	content, err := os.ReadFile(plmPath)
	if err != nil {
		return FileBlocks{}, err
	}
	// Parse blocks
	blocks, err := p.parseBlocks(string(content))
	if err != nil {
		return FileBlocks{}, err
	}
	// Create python content
	newContent := p.replaceBlocksInContent(string(content), blocks)

	// Write the .py file
	pyPath := plmPath + ".py"
	if err := os.WriteFile(pyPath, []byte(newContent), 0644); err != nil {
		return FileBlocks{}, err
	}

	return FileBlocks{
		FilePath: plmPath,
		Blocks:   blocks,
	}, nil
}

// updatePMLFileWithResults updates the PML file with results
func (p *Parser) updatePMLFileWithResults(blocks []Block, content string, results []string, localResultsDir string, sourceFile string) string {
	lines := strings.Split(content, "\n")
	var newLines []string
	var currentBlock *Block
	blockIndex := 0

	for _, line := range lines {
		trim := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trim, DirectiveAsk), strings.HasPrefix(trim, DirectiveDo):
			currentBlock = &Block{Type: trim}
			newLines = append(newLines, line)
		case strings.HasPrefix(trim, DirectiveEnd):
			if currentBlock != nil && blockIndex < len(results) {
				// Generate unique friendly name for this result
				resultName := p.generateUniqueResultName(sourceFile, blockIndex, localResultsDir)
				// Create result file in the local .pml directory
				resultFile := fmt.Sprintf("%s.pml", resultName)
				resultPath := filepath.Join(localResultsDir, resultFile)

				if err := p.writeResult(blocks[blockIndex], results[blockIndex], resultPath, localResultsDir, results[blockIndex]); err != nil {
					// If we can't write the result, just add it directly (fallback)
					newLines = append(newLines, line)
					newLines = append(newLines, results[blockIndex])
				} else {
					// Add link to result file with summary in the original format
					newLines = append(newLines, fmt.Sprintf(":--(r/%s:\"%s\")", resultName, results[blockIndex]))
				}
				blockIndex++
				currentBlock = nil
			} else {
				newLines = append(newLines, line)
			}
		default:
			newLines = append(newLines, line)
		}
	}

	return strings.Join(newLines, "\n")
}
