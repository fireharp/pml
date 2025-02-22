// Package parser implements the PML (Programming Language Markup) parser.
package parser

import (
	"context"
	"sync"
	"time"
)

// LLMClient interface for making LLM requests
type LLMClient interface {
	Ask(ctx context.Context, prompt string) (string, error)
	Summarize(ctx context.Context, text string) (string, error)
}

type Parser struct {
	llm            LLMClient
	sourcesDir     string
	compiledDir    string
	rootResultsDir string // For larger logs and detailed execution results
	cacheFile      string // Path to the cache file
	cache          map[string]CacheEntry
	cacheMu        sync.RWMutex // Protects cache map
	saveMu         sync.Mutex   // Protects cache file operations
	debug          bool
	forceProcess   bool
	resultFiles    sync.Map // Map to track result files being written
	fileLocks      sync.Map // Map to track file locks
	usedNamesMu    sync.Mutex
	usedNames      map[string]bool
}

// Block represents a block in PML file
type Block struct {
	Type        string
	Content     []string
	Response    string
	IsEphemeral bool // Whether this block was generated during runtime
	Start       int  // Start position in the original content
	End         int  // End position in the original content
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

// Directives used in PML files
const (
	DirectiveAsk = ":ask"
	DirectiveDo  = ":do"
	DirectiveEnd = ":--"
)

// Word lists for generating unique result names
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
